package celfilter

func Convert(expr string, opts ...Option) (string, error) {
	cvt := dftConverter
	if len(opts) != 0 {
		var err error
		cvt, err = cvt.Extend(opts...)
		if err != nil {
			return "", err
		}
	}
	return cvt.Convert(expr)
}

var dftConverter *rawConverter

func init() {
	dftConverter, _ = NewConverter()
}
