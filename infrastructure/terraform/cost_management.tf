# Budgets alert; they do not block AWS spend. The small thresholds make a
# surprise visible early while still allowing the normal billing pipeline to
# report the actual charge amount.
resource "aws_budgets_budget" "monthly_cost_2_usd" {
  name              = "${local.name_prefix}-monthly-cost-2-usd"
  budget_type       = "COST"
  limit_amount      = "2"
  limit_unit        = "USD"
  time_unit         = "MONTHLY"
  time_period_start = "2026-08-01_00:00"

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = [var.billing_alert_email]
  }
}

resource "aws_budgets_budget" "monthly_cost_5_usd" {
  name              = "${local.name_prefix}-monthly-cost-5-usd"
  budget_type       = "COST"
  limit_amount      = "5"
  limit_unit        = "USD"
  time_unit         = "MONTHLY"
  time_period_start = "2026-08-01_00:00"

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = [var.billing_alert_email]
  }
}

# AWS permits only one dimensional SERVICE monitor per account. The existing
# account-wide Default-Services-Monitor is external infrastructure, so this
# project references its ARN rather than importing or attempting to own it.
# DAILY supports direct email delivery; IMMEDIATE alerts require an SNS topic
# and are unnecessary for this low-cost portfolio.
resource "aws_ce_anomaly_subscription" "daily_email" {
  name      = "${local.name_prefix}-daily-cost-anomalies"
  frequency = "DAILY"

  monitor_arn_list = [var.service_anomaly_monitor_arn]

  subscriber {
    type    = "EMAIL"
    address = var.billing_alert_email
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      match_options = ["GREATER_THAN_OR_EQUAL"]
      values        = ["2"]
    }
  }
}
