package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
)

var ErrCorruptFile = errors.New("file payload is malformed or corrupted")

const (
	TargetPrintDPI   = 300
	MinimumAcceptDPI = 150
	PointsPerInch    = 72.0
)

type PreflightResult struct {
	Format          string
	WidthPts        float64
	HeightPts       float64
	MinEffectiveDPI int
	IsCorrupt       bool
	Warnings        []string
}

func InspectImage(data []byte, physicalWidthInches, physicalHeightInches float64) (*PreflightResult, error) {
	reader := bytes.NewReader(data)
	cfg, format, err := image.DecodeConfig(reader)
	if err != nil {
		return &PreflightResult{IsCorrupt: true}, ErrCorruptFile
	}

	result := &PreflightResult{
		Format:    format,
		WidthPts:  physicalWidthInches * PointsPerInch,
		HeightPts: physicalHeightInches * PointsPerInch,
	}

	dpiX := float64(cfg.Width) / physicalWidthInches
	dpiY := float64(cfg.Height) / physicalHeightInches
	minDPI := int(dpiX)
	if int(dpiY) < minDPI {
		minDPI = int(dpiY)
	}
	result.MinEffectiveDPI = minDPI

	if minDPI < MinimumAcceptDPI {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Image resolution is very low (%d DPI). Print will look blurry.", minDPI))
	} else if minDPI < TargetPrintDPI {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Image resolution (%d DPI) is below the recommended 300 DPI.", minDPI))
	}

	return result, nil
}
