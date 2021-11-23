package utils

import (
	"reflect"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

func IsSameFunc(fnL, fnR interface{}) bool {
	sfL := reflect.ValueOf(fnL)
	sfR := reflect.ValueOf(fnR)
	return sfL.Pointer() == sfR.Pointer()
}

func CallerPackageName(skipPkgName string) string {
	name := ""
	for i := 2; ; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			return ""
		}
		caller := runtime.FuncForPC(pc)
		log.Debugf("CallerPackageName get name=%s", caller.Name())
		name = getPkgNameByFuncName(caller.Name())
		log.Debugf("CallerPackageName get after name=%s", name)
		if name != skipPkgName {
			break
		}
	}
	return name
}

func GetPkgNameByFunc(fn interface{}) string {
	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return getPkgNameByFuncName(fnName)
}

func getPkgNameByFuncName(fnName string) string {
	pkgs := strings.Split(fnName, "/")
	fnName = pkgs[len(pkgs)-1]
	names := strings.Split(fnName, ".")
	return names[0]
}
