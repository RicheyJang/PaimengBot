package images

import "github.com/fogleman/gg"

// PasteStringDefault 以默认方式（默认字体、黑色、Wrapped、左上角定位、文字居左）贴文字
func (img *ImageCtx) PasteStringDefault(str string, fontSize, lineSpace float64, x, y, width float64) error {
	img.Push()
	defer img.Pop()
	if err := img.UseDefaultFont(fontSize); err != nil {
		return err
	} // 默认字体
	img.SetRGB(0, 0, 0) // 纯黑色
	img.DrawStringWrapped(str, x, y, 0, 0, width, lineSpace, gg.AlignLeft)
	return nil
}
