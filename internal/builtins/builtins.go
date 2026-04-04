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

func GetBuiltins() map[string]*values.BuiltInFunction {

	return map[string]*values.BuiltInFunction{

		// System functions
		"off":     values.NewBuiltInFunction("off", offFunction),
		"error":   values.NewBuiltInFunction("error", errorFunction),
		"args":    values.NewBuiltInFunction("args", argsFunction),
		"load":    values.NewBuiltInFunction("load", loadFunction),
		"loadopt": values.NewBuiltInFunction("loadopt", loadoptFunction),

		// FSIO
		"fopen":  values.NewBuiltInFunction("fopen", fopenFunction),
		"fclose": values.NewBuiltInFunction("fclose", fcloseFunction),
		"fwrite": values.NewBuiltInFunction("fwrite", fwriteFunction),
		"fread":  values.NewBuiltInFunction("fread", freadFunction),

		"getcwd": values.NewBuiltInFunction("getcwd", getcwdFunction),
		"chdir":  values.NewBuiltInFunction("chdir", chdirFunction),
		"stat":   values.NewBuiltInFunction("stat", statFunction),
		"lstat":  values.NewBuiltInFunction("lstat", lstatFunction),
		"unlink": values.NewBuiltInFunction("unlink", unlinkFunction),
		"mkdir":  values.NewBuiltInFunction("mkdir", mkdirFunction),
		"rename": values.NewBuiltInFunction("rename", renameFunction),
		"chmod":  values.NewBuiltInFunction("chmod", chmodFunction),
		"utime":  values.NewBuiltInFunction("utime", utimeFunction),
		"lutime": values.NewBuiltInFunction("lutime", lutimeFunction),

		"seek":  values.NewBuiltInFunction("seek", seekFunction),
		"fsync": values.NewBuiltInFunction("fsync", fsyncFunction),

		// Directory handling
		"dopen":  values.NewBuiltInFunction("dopen", dopenFunction),
		"dread":  values.NewBuiltInFunction("dread", dreadFunction),
		"dclose": values.NewBuiltInFunction("dclose", dcloseFunction),

		// Symlinks
		"symlink":  values.NewBuiltInFunction("symlink", symlinkFunction),
		"readlink": values.NewBuiltInFunction("readlink", readlinkFunction),

		// Type conversions
		"str":  values.NewBuiltInFunction("str", strFunction),
		"num":  values.NewBuiltInFunction("num", numFunction),
		"int":  values.NewBuiltInFunction("int", intFunction),
		"list": values.NewBuiltInFunction("list", listFunction),
		"type": values.NewBuiltInFunction("type", typeFunction),
		"len":  values.NewBuiltInFunction("len", lenFunction),
		"lenv": values.NewBuiltInFunction("lenv", lenvFunction),

		// Value checks
		"iserr": values.NewBuiltInFunction("iserr", iserrFunction),
		"isok":  values.NewBuiltInFunction("isok", isokFunction),

		// List operations
		"append": values.NewBuiltInFunction("append", appendFunction),
		"sort":   values.NewBuiltInFunction("sort", sortFunction),

		// Map operations
		"keys": values.NewBuiltInFunction("keys", keysFunction),
		"has":  values.NewBuiltInFunction("has", hasFunction),

		// Bytes
		"pack":   values.NewBuiltInFunction("pack", packFunction),
		"unpack": values.NewBuiltInFunction("unpack", unpackFunction),

		// Env
		"getenv": values.NewBuiltInFunction("getenv", getenvFunction),
		"setenv": values.NewBuiltInFunction("setenv", setenvFunction),

		// Sockets
		"sopen":   values.NewBuiltInFunction("sopen", sopenFunction),
		"sread":   values.NewBuiltInFunction("sread", sreadFunction),
		"swrite":  values.NewBuiltInFunction("swrite", swriteFunction),
		"sclose":  values.NewBuiltInFunction("sclose", scloseFunction),
		"saccept": values.NewBuiltInFunction("saccept", sacceptFunction),

		// Processes
		"popen":  values.NewBuiltInFunction("popen", popenFunction),
		"pclose": values.NewBuiltInFunction("pclose", pcloseFunction),
		"kill":   values.NewBuiltInFunction("kill", killFunction),
		"exec":   values.NewBuiltInFunction("exec", execFunction),
		"spawn":  values.NewBuiltInFunction("spawn", spawnFunction),

		// Randomness
		"rand": values.NewBuiltInFunction("rand", randFunction),

		// String operations
		"charat":  values.NewBuiltInFunction("charat", charatFunction),
		"slice":   values.NewBuiltInFunction("slice", sliceFunction),
		"replace": values.NewBuiltInFunction("replace", replaceFunction),
		"split":   values.NewBuiltInFunction("split", splitFunction),
		"indexof": values.NewBuiltInFunction("indexof", indexofFunction),
		"join":    values.NewBuiltInFunction("join", joinFunction),
		"lower":   values.NewBuiltInFunction("lower", lowerFunction),
		"upper":   values.NewBuiltInFunction("upper", upperFunction),

		// Math operations
		"sin": values.NewBuiltInFunction("sin", sinFunction),
		"cos": values.NewBuiltInFunction("cos", cosFunction),
		"tan": values.NewBuiltInFunction("tan", tanFunction),

		// Misc
		"getpid": values.NewBuiltInFunction("getpid", getpidFunction),
		"getuid": values.NewBuiltInFunction("getuid", getuidFunction),
		"time":   values.NewBuiltInFunction("time", timeFunction),
		"sleep":  values.NewBuiltInFunction("sleep", sleepFunction),
		"getch":  values.NewBuiltInFunction("getch", getchFunction),

		// Terminal state
		"getterm": values.NewBuiltInFunction("getterm", gettermFunction),
		"setterm": values.NewBuiltInFunction("setterm", settermFunction),
	}
}

func GetConstants() map[string]values.Value {
	return map[string]values.Value{

		"CHIPVR": values.NewString(constants.STR_LPLVR),
		"CHIPCN": values.NewString(constants.STR_LPLCN),
		"CHIPOS": values.NewString(constants.STR_LPLOS),
		"CHIPAR": values.NewString(constants.STR_LPLAR),

		"null":  values.NewNumber(constants.NUM_NUL),
		"false": values.NewNumber(constants.NUM_FAL),
		"true":  values.NewNumber(constants.NUM_TRU),
		"err":   values.NewString(constants.STR_ERR),
		"ok":    values.NewString(constants.STR_OK),
	}
}
