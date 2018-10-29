// Copyright Project Harbor Authors. All rights reserved.

package impl

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Colstuwjx/job/config"
	"github.com/Colstuwjx/job/env"
	"github.com/Colstuwjx/job/impl/job"
	jlogger "github.com/Colstuwjx/job/impl/logger"
	"github.com/Colstuwjx/job/logger"
)

const (
	maxRetryTimes = 5
)

// Context ...
type Context struct {
	// System context
	sysContext context.Context

	// Logger for job
	logger logger.Interface

	// op command func
	opCommandFunc job.CheckOPCmdFunc

	// checkin func
	checkInFunc job.CheckInFunc

	// other required information
	properties map[string]interface{}
}

// NewContext ...
func NewContext(sysCtx context.Context) *Context {
	return &Context{
		sysContext: sysCtx,
		properties: make(map[string]interface{}),
	}
}

// Init ...
func (c *Context) Init() error {
	// TODO: initial db etc

	return nil
}

// Build implements the same method in env.JobContext interface
// This func will build the job execution context before running
func (c *Context) Build(dep env.JobData) (env.JobContext, error) {
	jContext := &Context{
		sysContext: c.sysContext,
		properties: make(map[string]interface{}),
	}

	// Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	// Init logger here
	logPath := fmt.Sprintf("%s/%s.log", config.GetLogBasePath(), dep.ID)
	jContext.logger = jlogger.New(logPath, config.GetLogLevel())
	if jContext.logger == nil {
		return nil, errors.New("failed to initialize job logger")
	}

	if opCommandFunc, ok := dep.ExtraData["opCommandFunc"]; ok {
		if reflect.TypeOf(opCommandFunc).Kind() == reflect.Func {
			if funcRef, ok := opCommandFunc.(job.CheckOPCmdFunc); ok {
				jContext.opCommandFunc = funcRef
			}
		}
	}

	if jContext.opCommandFunc == nil {
		return nil, errors.New("failed to inject opCommandFunc")
	}

	if checkInFunc, ok := dep.ExtraData["checkInFunc"]; ok {
		if reflect.TypeOf(checkInFunc).Kind() == reflect.Func {
			if funcRef, ok := checkInFunc.(job.CheckInFunc); ok {
				jContext.checkInFunc = funcRef
			}
		}
	}

	if jContext.checkInFunc == nil {
		return nil, errors.New("failed to inject checkInFunc")
	}

	return jContext, nil
}

// Get implements the same method in env.JobContext interface
func (c *Context) Get(prop string) (interface{}, bool) {
	v, ok := c.properties[prop]
	return v, ok
}

// SystemContext implements the same method in env.JobContext interface
func (c *Context) SystemContext() context.Context {
	return c.sysContext
}

// Checkin is bridge func for reporting detailed status
func (c *Context) Checkin(status string) error {
	if c.checkInFunc != nil {
		c.checkInFunc(status)
	} else {
		return errors.New("nil check in function")
	}

	return nil
}

// OPCommand return the control operational command like stop/cancel if have
func (c *Context) OPCommand() (string, bool) {
	if c.opCommandFunc != nil {
		return c.opCommandFunc()
	}

	return "", false
}

// GetLogger returns the logger
func (c *Context) GetLogger() logger.Interface {
	return c.logger
}
