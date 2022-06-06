package utils

import (
	"fmt"
	"runtime"
	"strings"
)

func GetFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	f := runtime.FuncForPC(pc)
	fullName := f.Name()
	index := strings.LastIndexAny(fullName, "/\\")
	if index < 0 {
		return fullName
	}
	return fullName[index+1:]
}

func GetCurrentFuncName() string {
	return GetFuncName(2)
}

func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("[%s] %s:\n%w", GetFuncName(2), msg, err)
}

func WrapErrorf(err error, pattern string, objs ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(pattern, objs...)
	return fmt.Errorf("[%s] %s:\n%w", GetFuncName(2), msg, err)
}
