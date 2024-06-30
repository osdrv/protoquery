package protoquery

type EvalOption func(EvalContext)

func WithUseDefault(useDefault bool) EvalOption {
	return func(ctx EvalContext) {
		ctx.Options().UseDefault = useDefault
	}
}

func WithEnforceBool(enforceBool bool) EvalOption {
	return func(ctx EvalContext) {
		ctx.Options().EnforceBool = enforceBool
	}
}

type EvalOptions struct {
	// UseDefault is used to determine if the default value should be returned if
	// the protobuf message property is not set.
	UseDefault bool
	// EnforceBool is a flag indicating that instead of returning the actual property
	// value, the expression should check its presence in the context message.
	EnforceBool bool
}

type EvalContext interface {
	This() any
	Options() *EvalOptions
	Copy(opts ...EvalOption) EvalContext
}

type IndexedEvalContext interface {
	EvalContext
	Index() int
}

type EvalContextImpl struct {
	this any
	opts *EvalOptions
}

func NewEvalContext(this any, opts ...EvalOption) EvalContext {
	ctx := &EvalContextImpl{
		this: this,
		opts: &EvalOptions{},
	}
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

var _ EvalContext = (*EvalContextImpl)(nil)

func (ctx *EvalContextImpl) This() any {
	return ctx.this
}

func (ctx *EvalContextImpl) Options() *EvalOptions {
	return ctx.opts
}

func (ctx *EvalContextImpl) Copy(opts ...EvalOption) EvalContext {
	return NewEvalContext(ctx.This(), opts...)
}

type IndexedEvalContextImpl struct {
	EvalContext
	index int
}

var _ IndexedEvalContext = (*IndexedEvalContextImpl)(nil)

func NewIndexedEvalContext(this any, index int, opts ...EvalOption) IndexedEvalContext {
	return &IndexedEvalContextImpl{
		EvalContext: NewEvalContext(this, opts...),
		index:       index,
	}
}

func (ctx *IndexedEvalContextImpl) Index() int {
	return ctx.index
}

func (ctx *IndexedEvalContextImpl) Copy(opts ...EvalOption) EvalContext {
	return &IndexedEvalContextImpl{
		EvalContext: ctx.EvalContext.Copy(opts...),
		index:       ctx.index,
	}
}
