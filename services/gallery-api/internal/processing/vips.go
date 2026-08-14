package processing

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// VipsTransformer uses the libvips CLI included in the worker container. This
// avoids a CGO binding and its platform-specific linker surface while retaining
// libvips' streaming image pipeline and WebP encoder.
type VipsTransformer struct {
	Binary           string
	HeaderBinary     string
	MaxDecodedPixels int
}

func NewVipsTransformer() VipsTransformer {
	return VipsTransformer{Binary: "vips", HeaderBinary: "vipsheader", MaxDecodedPixels: MaxDecodedPixels}
}

func (transformer VipsTransformer) Render(ctx context.Context, inputPath, outputDirectory string, variants []Variant) (ImageSize, []RenderedVariant, error) {
	if transformer.Binary == "" {
		transformer.Binary = "vips"
	}
	if transformer.HeaderBinary == "" {
		transformer.HeaderBinary = "vipsheader"
	}
	if transformer.MaxDecodedPixels == 0 {
		transformer.MaxDecodedPixels = MaxDecodedPixels
	}

	width, err := transformer.headerValue(ctx, inputPath, "width")
	if err != nil {
		return ImageSize{}, nil, err
	}
	height, err := transformer.headerValue(ctx, inputPath, "height")
	if err != nil {
		return ImageSize{}, nil, err
	}
	if width < 1 || height < 1 || int64(width)*int64(height) > int64(transformer.MaxDecodedPixels) {
		return ImageSize{}, nil, fmt.Errorf("image dimensions %dx%d exceed decoded-pixel limit", width, height)
	}

	rendered := make([]RenderedVariant, 0, len(variants))
	for _, variant := range variants {
		if variant.Name == "" || variant.MaxWidth < 1 {
			return ImageSize{}, nil, errors.New("invalid derivative profile")
		}
		outputPath := path.Join(outputDirectory, variant.Name+".webp")
		// libvips selects the encoder from the file extension and accepts encoder
		// options in brackets on that output filename. Passing a generic --format
		// flag is not supported by the Alpine vips thumbnail CLI.
		outputTarget := outputPath + "[Q=80,strip]"
		// libvips thumbnail applies EXIF orientation by default. This release's
		// CLI exposes only --no-rotate, so do not pass a non-portable positive
		// flag here.
		if err := transformer.run(ctx, "thumbnail", inputPath, outputTarget, strconv.Itoa(variant.MaxWidth), "--size", "down"); err != nil {
			return ImageSize{}, nil, fmt.Errorf("render %s: %w", variant.Name, err)
		}
		rendered = append(rendered, RenderedVariant{Variant: variant, Path: outputPath})
	}

	// thumbnail's auto-rotation determines the rendered geometry. Read the large
	// WebP rather than trusting the original EXIF orientation.
	largePath := path.Join(outputDirectory, "large.webp")
	width, err = transformer.headerValue(ctx, largePath, "width")
	if err != nil {
		return ImageSize{}, nil, err
	}
	height, err = transformer.headerValue(ctx, largePath, "height")
	if err != nil {
		return ImageSize{}, nil, err
	}
	return ImageSize{Width: width, Height: height}, rendered, nil
}

func (transformer VipsTransformer) headerValue(ctx context.Context, inputPath, field string) (int, error) {
	output, err := transformer.headerOutput(ctx, "-f", field, inputPath)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", field, err)
	}
	value, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("parse %s %q: %w", field, output, err)
	}
	return value, nil
}

func (transformer VipsTransformer) headerOutput(ctx context.Context, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, transformer.HeaderBinary, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return "", errors.New(message)
	}
	return string(output), nil
}

func (transformer VipsTransformer) run(ctx context.Context, arguments ...string) error {
	_, err := transformer.output(ctx, arguments...)
	return err
}

func (transformer VipsTransformer) output(ctx context.Context, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, transformer.Binary, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return "", errors.New(message)
	}
	return string(output), nil
}
