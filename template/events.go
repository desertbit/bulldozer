/*
 *  Bulldozer Framework
 *  Copyright (C) DesertBit
 */

package template

import (
	"github.com/golang/glog"
)

const (
	onTemplateExecution         = "OnTemplateExec"
	onTemplateExecutionFinished = "OnTemplateExecF"
)

//#############################//
//### Template Struct Event ###//
//#############################//

// OnTemplateExecution is triggered during each template execution
func (t *Template) OnTemplateExecution(f func(c *Context, data interface{})) {
	t.emitter.On(onTemplateExecution, f)
}

// OnceTemplateExecution is the same event as OnTemplateExecution, but the listener is triggered only once
func (t *Template) OnceTemplateExecution(f func(c *Context, data interface{})) {
	t.emitter.Once(onTemplateExecution, f)
}

// OffTemplateExecution removes the listener again
func (t *Template) OffTemplateExecution(f func(c *Context, data interface{})) {
	t.emitter.Off(onTemplateExecution, f)
}

// Triggere the event
func (t *Template) triggerOnTemplateExecution(c *Context, data interface{}) {
	t.emitter.Emit(onTemplateExecution, c, data)
}

// OnTemplateExecutionFinished is triggered after each template execution
func (t *Template) OnTemplateExecutionFinished(f func(c *Context, data interface{})) {
	t.emitter.On(onTemplateExecutionFinished, f)
}

// OnceTemplateExecutionFinished is the same event as OnTemplateExecutionFinished, but the listener is triggered only once
func (t *Template) OnceTemplateExecutionFinished(f func(c *Context, data interface{})) {
	t.emitter.Once(onTemplateExecutionFinished, f)
}

// OffTemplateExecutionFinished removes the listener again
func (t *Template) OffTemplateExecutionFinished(f func(c *Context, data interface{})) {
	t.emitter.Off(onTemplateExecutionFinished, f)
}

// Triggere the event
func (t *Template) triggerOnTemplateExecutionFinished(c *Context, data interface{}) {
	t.emitter.Emit(onTemplateExecutionFinished, c, data)
}

//###############//
//### Private ###//
//###############//

func recoverEmitter(event interface{}, listener interface{}, err error) {
	glog.Errorf("bulldozer template events error: emitter event: %v: listener: %v: %v", event, listener, err)
}
