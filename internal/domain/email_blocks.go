package domain

type EmailBlock struct {
	ID       string       `json:"id"`
	Kind     string       `json:"kind"`
	Path     string       `json:"path"`
	Children []EmailBlock `json:"children"`
	Data     any          `json:"data"`
}

// GetBlockData returns the block data cast to the appropriate type based on the Kind
func (n *EmailBlock) GetBlockData() any {
	switch n.Kind {
	case "button":
		return n.Data.(ButtonBlockData)
	case "image":
		return n.Data.(ImageBlockData)
	case "column":
		return n.Data.(ColumnBlockData)
	case "divider":
		return n.Data.(DividerBlockData)
	case "section":
		return n.Data.(SectionBlockData)
	case "openTracking":
		return n.Data.(OpenTrackingBlockData)
	case "text":
		return n.Data.(TextBlockData)
	default:
		return n.Data
	}
}

// BaseStyles represents common style properties used across multiple blocks
type BaseStyles struct {
	PaddingTop    string `json:"paddingTop,omitempty"`
	PaddingRight  string `json:"paddingRight,omitempty"`
	PaddingBottom string `json:"paddingBottom,omitempty"`
	PaddingLeft   string `json:"paddingLeft,omitempty"`
	Padding       string `json:"padding,omitempty"`
	BorderStyle   string `json:"borderStyle,omitempty"`
	BorderWidth   string `json:"borderWidth,omitempty"`
	BorderColor   string `json:"borderColor,omitempty"`
	BorderRadius  string `json:"borderRadius,omitempty"`
}

// WrapperStyles represents common wrapper properties for blocks
type WrapperStyles struct {
	Align          string `json:"align"`
	PaddingControl string `json:"paddingControl"` // "all" or "separate"
	Padding        string `json:"padding,omitempty"`
	PaddingTop     string `json:"paddingTop,omitempty"`
	PaddingRight   string `json:"paddingRight,omitempty"`
	PaddingBottom  string `json:"paddingBottom,omitempty"`
	PaddingLeft    string `json:"paddingLeft,omitempty"`
}

// ButtonBlockData represents the data structure for a button block
type ButtonBlockData struct {
	Button struct {
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
		BorderControl          string `json:"borderControl"` // "all" or "separate"
	} `json:"button"`
	Wrapper WrapperStyles `json:"wrapper"`
}

// ImageBlockData represents the data structure for an image block
type ImageBlockData struct {
	Image struct {
		Src           string `json:"src"`
		Alt           string `json:"alt"`
		Href          string `json:"href"`
		Width         string `json:"width"`
		BorderControl string `json:"borderControl"` // "all" or "separate"
	} `json:"image"`
	Wrapper WrapperStyles `json:"wrapper"`
}

// ColumnBlockData represents the data structure for a column block
type ColumnBlockData struct {
	Styles struct {
		VerticalAlign   string `json:"verticalAlign"` // "top", "middle", "bottom"
		BackgroundColor string `json:"backgroundColor,omitempty"`
		MinHeight       string `json:"minHeight,omitempty"`
		BaseStyles
	} `json:"styles"`
	PaddingControl string `json:"paddingControl"` // "all" or "separate"
	BorderControl  string `json:"borderControl"`  // "all" or "separate"
}

// DividerBlockData represents the data structure for a divider block
type DividerBlockData struct {
	Align           string `json:"align"` // "left", "center", "right"
	BorderColor     string `json:"borderColor"`
	BorderStyle     string `json:"borderStyle"`
	BorderWidth     string `json:"borderWidth"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Width           string `json:"width"`
	PaddingControl  string `json:"paddingControl"` // "all" or "separate"
	Padding         string `json:"padding,omitempty"`
	PaddingTop      string `json:"paddingTop,omitempty"`
	PaddingRight    string `json:"paddingRight,omitempty"`
	PaddingBottom   string `json:"paddingBottom,omitempty"`
	PaddingLeft     string `json:"paddingLeft,omitempty"`
}

// SectionBlockData represents the data structure for a section block
type SectionBlockData struct {
	ColumnsOnMobile     bool   `json:"columnsOnMobile"`
	StackColumnsAtWidth int    `json:"stackColumnsAtWidth"`
	BackgroundType      string `json:"backgroundType"` // "color" or "image"
	PaddingControl      string `json:"paddingControl"` // "all" or "separate"
	BorderControl       string `json:"borderControl"`  // "all" or "separate"
	Styles              struct {
		TextAlign        string `json:"textAlign"`                  // "left", "center", "right", "justify"
		BackgroundRepeat string `json:"backgroundRepeat,omitempty"` // "repeat", "no-repeat", "repeat-x", "repeat-y"
		Padding          string `json:"padding,omitempty"`
		BorderWidth      string `json:"borderWidth,omitempty"`
		BorderStyle      string `json:"borderStyle,omitempty"`
		BorderColor      string `json:"borderColor,omitempty"`
		BackgroundColor  string `json:"backgroundColor,omitempty"`
		BackgroundImage  string `json:"backgroundImage,omitempty"`
		BackgroundSize   string `json:"backgroundSize,omitempty"` // "cover" or "contain"
		BaseStyles
	} `json:"styles"`
}

// OpenTrackingBlockData represents the data structure for an open tracking block
type OpenTrackingBlockData struct {
	// No specific data structure
}

// TextBlockData represents the data structure for a text block
type TextBlockData struct {
	Align           string `json:"align"` // "left", "center", "right"
	Width           string `json:"width"`
	HyperlinkStyles struct {
		Color          string `json:"color"`
		TextDecoration string `json:"textDecoration"`
		FontFamily     string `json:"fontFamily"`
		FontSize       string `json:"fontSize"`
		FontWeight     int    `json:"fontWeight"`
		FontStyle      string `json:"fontStyle"`
		TextTransform  string `json:"textTransform"`
	} `json:"hyperlinkStyles"`
	EditorData []struct {
		Type     string `json:"type"`
		Children []struct {
			Text string `json:"text"`
		} `json:"children"`
	} `json:"editorData"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	PaddingControl  string `json:"paddingControl,omitempty"` // "all" or "separate"
	Padding         string `json:"padding,omitempty"`
	PaddingTop      string `json:"paddingTop,omitempty"`
	PaddingRight    string `json:"paddingRight,omitempty"`
	PaddingBottom   string `json:"paddingBottom,omitempty"`
	PaddingLeft     string `json:"paddingLeft,omitempty"`
}

// HeadingBlockData represents the data structure for a heading block
type HeadingBlockData struct {
	Type       string `json:"type"`  // "h1", "h2", "h3"
	Align      string `json:"align"` // "left", "center", "right"
	Width      string `json:"width"`
	EditorData []struct {
		Type     string `json:"type"`
		Children []struct {
			Text string `json:"text"`
		} `json:"children"`
	} `json:"editorData"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	PaddingControl  string `json:"paddingControl,omitempty"` // "all" or "separate"
	Padding         string `json:"padding,omitempty"`
	PaddingTop      string `json:"paddingTop,omitempty"`
	PaddingRight    string `json:"paddingRight,omitempty"`
	PaddingBottom   string `json:"paddingBottom,omitempty"`
	PaddingLeft     string `json:"paddingLeft,omitempty"`
}

// LiquidBlockData represents the data structure for a liquid template block
type LiquidBlockData struct {
	LiquidCode string `json:"liquidCode"`
}

// RootBlockData represents the data structure for a root block
type RootBlockData struct {
	Styles struct {
		Body struct {
			Width           string `json:"width"`
			Margin          string `json:"margin"`
			BackgroundColor string `json:"backgroundColor"`
		} `json:"body"`
		H1 struct {
			Color          string `json:"color"`
			FontSize       string `json:"fontSize"`
			FontStyle      string `json:"fontStyle"`
			FontWeight     int    `json:"fontWeight"`
			PaddingControl string `json:"paddingControl"` // "all" or "separate"
			Padding        string `json:"padding,omitempty"`
			PaddingTop     string `json:"paddingTop,omitempty"`
			PaddingRight   string `json:"paddingRight,omitempty"`
			PaddingBottom  string `json:"paddingBottom,omitempty"`
			PaddingLeft    string `json:"paddingLeft,omitempty"`
			Margin         string `json:"margin"`
			FontFamily     string `json:"fontFamily"`
		} `json:"h1"`
		H2 struct {
			Color          string `json:"color"`
			FontSize       string `json:"fontSize"`
			FontStyle      string `json:"fontStyle"`
			FontWeight     int    `json:"fontWeight"`
			PaddingControl string `json:"paddingControl"` // "all" or "separate"
			Padding        string `json:"padding,omitempty"`
			PaddingTop     string `json:"paddingTop,omitempty"`
			PaddingRight   string `json:"paddingRight,omitempty"`
			PaddingBottom  string `json:"paddingBottom,omitempty"`
			PaddingLeft    string `json:"paddingLeft,omitempty"`
			Margin         string `json:"margin"`
			FontFamily     string `json:"fontFamily"`
		} `json:"h2"`
		H3 struct {
			Color          string `json:"color"`
			FontSize       string `json:"fontSize"`
			FontStyle      string `json:"fontStyle"`
			FontWeight     int    `json:"fontWeight"`
			PaddingControl string `json:"paddingControl"` // "all" or "separate"
			Padding        string `json:"padding,omitempty"`
			PaddingTop     string `json:"paddingTop,omitempty"`
			PaddingRight   string `json:"paddingRight,omitempty"`
			PaddingBottom  string `json:"paddingBottom,omitempty"`
			PaddingLeft    string `json:"paddingLeft,omitempty"`
			Margin         string `json:"margin"`
			FontFamily     string `json:"fontFamily"`
		} `json:"h3"`
		Paragraph struct {
			Color          string `json:"color"`
			FontSize       string `json:"fontSize"`
			FontStyle      string `json:"fontStyle"`
			FontWeight     int    `json:"fontWeight"`
			PaddingControl string `json:"paddingControl"` // "all" or "separate"
			Padding        string `json:"padding,omitempty"`
			PaddingTop     string `json:"paddingTop,omitempty"`
			PaddingRight   string `json:"paddingRight,omitempty"`
			PaddingBottom  string `json:"paddingBottom,omitempty"`
			PaddingLeft    string `json:"paddingLeft,omitempty"`
			Margin         string `json:"margin"`
			FontFamily     string `json:"fontFamily"`
		} `json:"paragraph"`
		Hyperlink struct {
			Color          string `json:"color"`
			TextDecoration string `json:"textDecoration"`
			FontFamily     string `json:"fontFamily"`
			FontSize       string `json:"fontSize"`
			FontWeight     int    `json:"fontWeight"`
			FontStyle      string `json:"fontStyle"`
			TextTransform  string `json:"textTransform"`
		} `json:"hyperlink"`
	} `json:"styles"`
}

// ColumnLayoutBlockData represents the base data structure for column layouts
type ColumnLayoutBlockData struct {
	SectionBlockData
	Columns []int `json:"columns"`
}

// OneColumnBlockData represents the data structure for a single column layout
type OneColumnBlockData struct {
	SectionBlockData
	Columns [1]int `json:"columns"` // [24]
}

// Column variations
type Columns168BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [16, 8]
}

type Columns204BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [20, 4]
}

type Columns420BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [4, 20]
}

type Columns816BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [8, 16]
}

type Columns888BlockData struct {
	ColumnLayoutBlockData
	Columns [3]int `json:"columns"` // [8, 8, 8]
}

type Columns1212BlockData struct {
	ColumnLayoutBlockData
	Columns [2]int `json:"columns"` // [12, 12]
}

type Columns6666BlockData struct {
	ColumnLayoutBlockData
	Columns [4]int `json:"columns"` // [6, 6, 6, 6]
}
