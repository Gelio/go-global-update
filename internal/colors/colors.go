package colors

import "github.com/fatih/color"

type DecoratorFactory struct {
	// forceColors forces printing colors in the output even if the standard
	// output is not a TTY.
	forceColors bool
}

func NewFactory(forceColors bool) DecoratorFactory {
	return DecoratorFactory{
		forceColors,
	}
}

// NewDecorator returns a Sprintf-like function that applies the attributes to
// the formatted string.
func (f *DecoratorFactory) NewDecorator(attributes ...color.Attribute) func(a ...interface{}) string {
	c := color.New(attributes...)
	if f.forceColors {
		c.EnableColor()
	}

	return c.Sprint
}
