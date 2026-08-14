# The processing queue is intentionally small and durable. A batch size of one
# in the Lambda mapping (below) makes a failed image retry independently rather
# than causing an otherwise successful upload to be replayed.
resource "aws_sqs_queue" "gallery_image_processing_dlq" {
  name                      = "${local.image_worker_queue_name}-dlq"
  message_retention_seconds = 1209600

  tags = local.tags
}

resource "aws_sqs_queue" "gallery_image_processing" {
  name                       = local.image_worker_queue_name
  delay_seconds              = 15
  message_retention_seconds  = 345600
  visibility_timeout_seconds = var.image_worker_timeout_seconds * 6

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.gallery_image_processing_dlq.arn
    maxReceiveCount     = 3
  })

  tags = local.tags
}

# Only notifications from this originals bucket may send work to the queue.
data "aws_iam_policy_document" "gallery_image_processing_queue" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    actions   = ["sqs:SendMessage"]
    resources = [aws_sqs_queue.gallery_image_processing.arn]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [aws_s3_bucket.gallery_originals.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

resource "aws_sqs_queue_policy" "gallery_image_processing" {
  queue_url = aws_sqs_queue.gallery_image_processing.id
  policy    = data.aws_iam_policy_document.gallery_image_processing_queue.json
}
