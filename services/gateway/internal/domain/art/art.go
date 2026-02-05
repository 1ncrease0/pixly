package art

type SavePixelartInput struct {
	UserID  string
	Title   string
	Palette []string
	Pixels  []int64
	Width   int64
	Height  int64
}

type UpdateCanvasInput struct {
	UserID     string
	PixelartID string
	Palette    []string
	Pixels     []int64
	Width      int64
	Height     int64
}

type DeletePixelartInput struct {
	UserID     string
	PixelartID string
}

type GetUserPixelartInput struct {
	UserID     string
	PixelartID string
}

type Pixelart struct {
	Title    string
	Palette  []string
	Pixels   []int64
	Width    int64
	Height   int64
	ImageURL string
}

type GetUserPreviewsInput struct {
	UserID string
}

type Preview struct {
	PixelartID string
	Title      string
	ImageURL   string
}
