/*
 *
 * RR2 - internal/builtins/builtins.go
 *
 */

package builtins

import (
	"chip-go/internal/constants"
	"chip-go/internal/values"
)

func GetBuiltins() map[string]*values.NativeFunction {

	return map[string]*values.NativeFunction{
		// System functions
		"off":   values.NewNativeFunction("off", offFunction, values.BuiltIn),
		"error": values.NewNativeFunction("error", errorFunction, values.BuiltIn),
		"args":  values.NewNativeFunction("args", argsFunction, values.BuiltIn),
		"load":  values.NewNativeFunction("load", loadFunction, values.BuiltIn),

		// Plugins
		"pload": values.NewNativeFunction("pload", ploadFunction, values.BuiltIn),
		"plist": values.NewNativeFunction("plist", plistFunction, values.BuiltIn),

		// FSIO
		"fopen":  values.NewNativeFunction("fopen", fopenFunction, values.BuiltIn),
		"fclose": values.NewNativeFunction("fclose", fcloseFunction, values.BuiltIn),
		"fwrite": values.NewNativeFunction("fwrite", fwriteFunction, values.BuiltIn),
		"fread":  values.NewNativeFunction("fread", freadFunction, values.BuiltIn),

		"getcwd": values.NewNativeFunction("getcwd", getcwdFunction, values.BuiltIn),
		"chdir":  values.NewNativeFunction("chdir", chdirFunction, values.BuiltIn),
		"stat":   values.NewNativeFunction("stat", statFunction, values.BuiltIn),
		"lstat":  values.NewNativeFunction("lstat", lstatFunction, values.BuiltIn),
		"unlink": values.NewNativeFunction("unlink", unlinkFunction, values.BuiltIn),
		"mkdir":  values.NewNativeFunction("mkdir", mkdirFunction, values.BuiltIn),
		"rename": values.NewNativeFunction("rename", renameFunction, values.BuiltIn),
		"chmod":  values.NewNativeFunction("chmod", chmodFunction, values.BuiltIn),

		"seek":   values.NewNativeFunction("seek", seekFunction, values.BuiltIn),
		"fsync":  values.NewNativeFunction("fsync", fsyncFunction, values.BuiltIn),
		"select": values.NewNativeFunction("select", selectFunction, values.BuiltIn),

		// Directory handling
		"dopen":  values.NewNativeFunction("dopen", dopenFunction, values.BuiltIn),
		"dread":  values.NewNativeFunction("dread", dreadFunction, values.BuiltIn),
		"dclose": values.NewNativeFunction("dclose", dcloseFunction, values.BuiltIn),

		// Symlinks
		"symlink":  values.NewNativeFunction("symlink", symlinkFunction, values.BuiltIn),
		"readlink": values.NewNativeFunction("readlink", readlinkFunction, values.BuiltIn),

		// Type conversions
		"str":  values.NewNativeFunction("str", strFunction, values.BuiltIn),
		"num":  values.NewNativeFunction("num", numFunction, values.BuiltIn),
		"int":  values.NewNativeFunction("int", intFunction, values.BuiltIn),
		"list": values.NewNativeFunction("list", listFunction, values.BuiltIn),
		"type": values.NewNativeFunction("type", typeFunction, values.BuiltIn),
		"len":  values.NewNativeFunction("len", lenFunction, values.BuiltIn),
		"lenv": values.NewNativeFunction("lenv", lenvFunction, values.BuiltIn),

		// List operations
		"append": values.NewNativeFunction("append", appendFunction, values.BuiltIn),
		"sort":   values.NewNativeFunction("sort", sortFunction, values.BuiltIn),

		// Bytes
		"pack":   values.NewNativeFunction("pack", packFunction, values.BuiltIn),
		"unpack": values.NewNativeFunction("unpack", unpackFunction, values.BuiltIn),

		// Env
		"getenv": values.NewNativeFunction("getenv", getenvFunction, values.BuiltIn),
		"setenv": values.NewNativeFunction("setenv", setenvFunction, values.BuiltIn),

		// Sockets
		"sopen":   values.NewNativeFunction("sopen", sopenFunction, values.BuiltIn),
		"sread":   values.NewNativeFunction("sread", sreadFunction, values.BuiltIn),
		"swrite":  values.NewNativeFunction("swrite", swriteFunction, values.BuiltIn),
		"sclose":  values.NewNativeFunction("sclose", scloseFunction, values.BuiltIn),
		"saccept": values.NewNativeFunction("saccept", sacceptFunction, values.BuiltIn),

		// Processes
		"popen":  values.NewNativeFunction("popen", popenFunction, values.BuiltIn),
		"pclose": values.NewNativeFunction("pclose", pcloseFunction, values.BuiltIn),
		"kill":   values.NewNativeFunction("kill", killFunction, values.BuiltIn),
		"exec":   values.NewNativeFunction("exec", execFunction, values.BuiltIn),

		// Randomness
		"rand": values.NewNativeFunction("rand", randFunction, values.BuiltIn),

		// String operations
		"charat":  values.NewNativeFunction("charat", charatFunction, values.BuiltIn),
		"substr":  values.NewNativeFunction("substr", substrFunction, values.BuiltIn),
		"replace": values.NewNativeFunction("replace", replaceFunction, values.BuiltIn),
		"split":   values.NewNativeFunction("split", splitFunction, values.BuiltIn),
		"indexof": values.NewNativeFunction("indexof", indexofFunction, values.BuiltIn),
		"join":    values.NewNativeFunction("join", joinFunction, values.BuiltIn),
		"lower":   values.NewNativeFunction("lower", lowerFunction, values.BuiltIn),
		"upper":   values.NewNativeFunction("upper", upperFunction, values.BuiltIn),

		// Math operations
		"sin": values.NewNativeFunction("sin", sinFunction, values.BuiltIn),
		"cos": values.NewNativeFunction("cos", cosFunction, values.BuiltIn),
		"tan": values.NewNativeFunction("tan", tanFunction, values.BuiltIn),

		// Misc
		"getpid": values.NewNativeFunction("getpid", getpidFunction, values.BuiltIn),
		"getuid": values.NewNativeFunction("getuid", getuidFunction, values.BuiltIn),
		"time":   values.NewNativeFunction("time", timeFunction, values.BuiltIn),
		"sleep":  values.NewNativeFunction("sleep", sleepFunction, values.BuiltIn),
		"getch":  values.NewNativeFunction("getch", getchFunction, values.BuiltIn),
	}
}

func GetConstants() map[string]values.Value {
	return map[string]values.Value{
		"null":  values.NewNumber(constants.NUM_NUL),
		"false": values.NewNumber(constants.NUM_FAL),
		"true":  values.NewNumber(constants.NUM_TRU),
		"err":   values.NewString(constants.STR_ERR),
		"ok":    values.NewString(constants.STR_OK),

		"CHIPVR": values.NewString(constants.STR_LPLVR),
		"CHIPCN": values.NewString(constants.STR_LPLCN),
		"CHIPOS": values.NewString(constants.STR_LPLOS),
		"CHIPAR": values.NewString(constants.STR_LPLAR),

		"BUNDLERUN": values.NewNumber(constants.NUM_NUL),
		"BUNDLEDIR": values.NewString(constants.STR_ERR),
		"BUNDLEID":  values.NewNumber(constants.NUM_NUL),
	}
}
