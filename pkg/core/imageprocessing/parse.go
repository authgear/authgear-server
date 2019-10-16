package imageprocessing

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses the query into a series of operations.
func Parse(query string) ([]Operation, error) {
	parts := strings.Split(query, "/")

	assetTypeStr := parts[0]
	_, err := parseAssetType(assetTypeStr)
	if err != nil {
		return nil, err
	}

	var ops []Operation
	rest := parts[1:]
	for _, part := range rest {
		op, err := parseOperation(part)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func parseAssetType(s string) (AssetType, error) {
	if AssetType(s) == AssetTypeImage {
		return AssetTypeImage, nil
	}
	return "", fmt.Errorf("invalid asset type: %v", s)
}

func parseOperation(s string) (Operation, error) {
	parts := strings.SplitN(s, ",", 2)
	opName := parts[0]
	rest := ""
	if len(parts) == 2 {
		rest = parts[1]
	}

	switch opName {
	case "format":
		return parseFormat(rest)
	case "quality":
		return parseQuality(rest)
	case "resize":
		return parseResize(rest)
	default:
		return nil, fmt.Errorf("invalid operation: %v", s)
	}
}

func parseFormat(s string) (*Format, error) {
	switch s {
	case string(ImageFormatJPEG):
		return &Format{
			ImageFormat: ImageFormatJPEG,
		}, nil
	case string(ImageFormatPNG):
		return &Format{
			ImageFormat: ImageFormatPNG,
		}, nil
	case string(ImageFormatWebP):
		return &Format{
			ImageFormat: ImageFormatWebP,
		}, nil
	default:
		return nil, fmt.Errorf("invalid format: %v", s)
	}
}

func parseQuality(s string) (*Quality, error) {
	name, valueStr := parseArg(s)
	if name != "Q" {
		return nil, fmt.Errorf("invalid quality: %v", s)
	}
	value, err := parseInt(valueStr, 1, 100)
	if err != nil {
		return nil, err
	}
	return &Quality{
		AbsoluteQuality: value,
	}, nil
}

func parseResize(s string) (*Resize, error) {
	resize := NewResize()
	parts := strings.Split(s, ",")
	for _, part := range parts {
		name, valueStr := parseArg(part)
		switch name {
		case "m":
			scalingMode, err := parseResizeScalingMode(valueStr)
			if err != nil {
				return nil, err
			}
			resize.ScalingMode = scalingMode
		case "w":
			w, err := parseInt(valueStr, 1, 4096)
			if err != nil {
				return nil, err
			}
			resize.Width = w
		case "h":
			h, err := parseInt(valueStr, 1, 4096)
			if err != nil {
				return nil, err
			}
			resize.Height = h
		case "l":
			l, err := parseInt(valueStr, 1, 4096)
			if err != nil {
				return nil, err
			}
			resize.LongerSide = l
		case "s":
			s, err := parseInt(valueStr, 1, 4096)
			if err != nil {
				return nil, err
			}
			resize.ShorterSide = s
		case "color":
			color, err := parseColor(valueStr)
			if err != nil {
				return nil, err
			}
			resize.Color = *color
		}
	}
	return resize, nil
}

func parseArg(arg string) (string, string) {
	parts := strings.SplitN(arg, "_", 2)
	name := parts[0]
	var valueStr string
	if len(parts) == 2 {
		valueStr = parts[1]
	}
	return name, valueStr
}

func parseInt(s string, min int, max int) (int, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("value '%s' is not an integer", s)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("value '%s' is not in range [%v,%v]", s, min, max)
	}
	return value, nil
}

func parseResizeScalingMode(s string) (ResizeScalingMode, error) {
	switch s {
	case string(ResizeScalingModeLfit):
		return ResizeScalingModeLfit, nil
	case string(ResizeScalingModeMfit):
		return ResizeScalingModeMfit, nil
	case string(ResizeScalingModePad):
		return ResizeScalingModePad, nil
	case string(ResizeScalingModeFixed):
		return ResizeScalingModeFixed, nil
	default:
		return "", fmt.Errorf("invalid scaling mode: %v", s)
	}
}

func parseColor(s string) (*Color, error) {
	if len(s) != 6 {
		return nil, fmt.Errorf("invalid color: %v", s)
	}
	parseHex := func(hex string) (int, error) {
		i, err := strconv.ParseInt(hex, 16, 0)
		if err != nil {
			return 0, fmt.Errorf("invalid color: %v", s)
		}
		return int(i), nil
	}
	r, err := parseHex(s[0:2])
	if err != nil {
		return nil, err
	}
	g, err := parseHex(s[2:4])
	if err != nil {
		return nil, err
	}
	b, err := parseHex(s[4:6])
	if err != nil {
		return nil, err
	}
	return &Color{
		R: r,
		G: g,
		B: b,
	}, nil
}
