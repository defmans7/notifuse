package domain

import (
	"reflect"
	"testing"
)

func TestEmailBlock_GetBlockData(t *testing.T) {
	tests := []struct {
		name     string
		block    EmailBlock
		expected any
		typeOf   string
	}{
		{
			name: "Button block",
			block: EmailBlock{
				Kind: "button",
				Data: ButtonBlockData{
					Button: struct {
						Text                   string `json:"text"`
						Href                   string `json:"href"`
						BackgroundColor        string `json:"backgroundColor"`
						FontFamily             string `json:"fontFamily"`
						FontSize               string `json:"fontSize"`
						FontWeight             int    `json:"fontWeight"`
						FontStyle              string `json:"fontStyle"`
						Color                  string `json:"color"`
						InnerVerticalPadding   string `json:"innerVerticalPadding"`
						InnerHorizontalPadding string `json:"innerHorizontalPadding"`
						Width                  string `json:"width"`
						TextTransform          string `json:"textTransform"`
						BorderRadius           string `json:"borderRadius"`
						DisableTracking        bool   `json:"disable_tracking"`
						BorderControl          string `json:"borderControl"`
					}{
						Text: "Click me",
					},
				},
			},
			typeOf: "domain.ButtonBlockData",
		},
		{
			name: "Image block",
			block: EmailBlock{
				Kind: "image",
				Data: ImageBlockData{
					Image: struct {
						Src           string `json:"src"`
						Alt           string `json:"alt"`
						Href          string `json:"href"`
						Width         string `json:"width"`
						BorderControl string `json:"borderControl"`
					}{
						Src: "https://example.com/image.jpg",
					},
				},
			},
			typeOf: "domain.ImageBlockData",
		},
		{
			name: "Column block",
			block: EmailBlock{
				Kind: "column",
				Data: ColumnBlockData{},
			},
			typeOf: "domain.ColumnBlockData",
		},
		{
			name: "Divider block",
			block: EmailBlock{
				Kind: "divider",
				Data: DividerBlockData{},
			},
			typeOf: "domain.DividerBlockData",
		},
		{
			name: "Section block",
			block: EmailBlock{
				Kind: "section",
				Data: SectionBlockData{},
			},
			typeOf: "domain.SectionBlockData",
		},
		{
			name: "OpenTracking block",
			block: EmailBlock{
				Kind: "openTracking",
				Data: OpenTrackingBlockData{},
			},
			typeOf: "domain.OpenTrackingBlockData",
		},
		{
			name: "Text block",
			block: EmailBlock{
				Kind: "text",
				Data: TextBlockData{},
			},
			typeOf: "domain.TextBlockData",
		},
		{
			name: "Unknown block type",
			block: EmailBlock{
				Kind: "unknown",
				Data: map[string]interface{}{"key": "value"},
			},
			typeOf: "map[string]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.block.GetBlockData()

			// Check that the result is of the expected type
			resultType := reflect.TypeOf(result).String()
			if resultType != tt.typeOf {
				t.Errorf("GetBlockData() returned incorrect type = %v, want %v", resultType, tt.typeOf)
			}

			// Additionally, check that data was properly cast by comparing with original
			switch tt.block.Kind {
			case "button":
				buttonData, ok := result.(ButtonBlockData)
				if !ok {
					t.Errorf("Failed to cast result to ButtonBlockData")
				}
				original := tt.block.Data.(ButtonBlockData)
				if buttonData.Button.Text != original.Button.Text {
					t.Errorf("Button text = %v, want %v", buttonData.Button.Text, original.Button.Text)
				}
			case "image":
				imageData, ok := result.(ImageBlockData)
				if !ok {
					t.Errorf("Failed to cast result to ImageBlockData")
				}
				original := tt.block.Data.(ImageBlockData)
				if imageData.Image.Src != original.Image.Src {
					t.Errorf("Image src = %v, want %v", imageData.Image.Src, original.Image.Src)
				}
			case "unknown":
				// For unknown types, the original data should be returned as-is
				mapData, ok := result.(map[string]interface{})
				if !ok {
					t.Errorf("Failed to cast result to map[string]interface{}")
				}
				if mapData["key"] != "value" {
					t.Errorf("Map data = %v, want %v", mapData["key"], "value")
				}
			}
		})
	}
}
