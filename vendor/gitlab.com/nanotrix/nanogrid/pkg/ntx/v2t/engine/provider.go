package engine

type (
	LabelToContextParser interface {
		Parse(label string) (*EngineContext, error)
	}

	LabelToContextParserFunc func(label string) (*EngineContext, error)

	labelToContext struct {
	}
)

func NewLabelToContextParser() LabelToContextParser {
	return &labelToContext{}
}

func (p *labelToContext) Parse(label string) (*EngineContext, error) {
	mods, err := ModulesFromLabel(label)
	if err != nil {
		return nil, err
	}

	ctx, err := ModuleToEngineContext(mods)
	if err != nil {
		return nil, err
	}

	ctx.AudioFormat = &AudioFormat{Formats: &AudioFormat_Auto{Auto: &AudioFormat_AutoDetect{ProbeSizeBytes: 0}}}

	return ctx, ctx.Valid()
}

func (f LabelToContextParserFunc) Parse(label string) (*EngineContext, error) {
	return f(label)
}
