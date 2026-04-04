" Vim syntax file
" Language: ChipLang
" Maintainer: ideumi
" Latest Revision: 2026-04-04

if exists("b:current_syntax")
  finish
endif

" Keywords
syn keyword chipKeyword var func return continue break if elseif else for while to step b m
syn keyword chipOperator and or not xor
syn keyword chipOperator band bor bnot bxor

" Constants
syn keyword chipConstant null true false CHIPVR CHIPCN CHIPOS CHIPAR
syn keyword chipError err
syn keyword chipSuccess ok

" Built-in functions
syn keyword chipBuiltin append args charat chdir chmod cos
syn keyword chipBuiltin dclose dopen dread
syn keyword chipBuiltin error exec
syn keyword chipBuiltin fclose fopen fread fsync fwrite
syn keyword chipBuiltin getch getcwd getenv getpid getterm getuid
syn keyword chipBuiltin has indexof int iserr isok
syn keyword chipBuiltin join
syn keyword chipBuiltin keys kill
syn keyword chipBuiltin len lenv list load loadopt lower lstat lutime
syn keyword chipBuiltin mkdir
syn keyword chipBuiltin num
syn keyword chipBuiltin off
syn keyword chipBuiltin pack pclose popen
syn keyword chipBuiltin rand readlink rename replace
syn keyword chipBuiltin saccept sclose seek setenv setterm sin sleep slice sopen sort spawn split sread stat str swrite symlink
syn keyword chipBuiltin tan time type
syn keyword chipBuiltin unlink unpack upper utime

" Comments
syn match chipComment "#.*$" contains=chipTodo
syn keyword chipTodo contained TODO FIXME NOTE WARNING HACK BUG XXX

" Strings
syn region chipString start='"' end='"' skip='\\"' contains=chipStringEscape
syn match chipStringEscape contained '\\[nrt"\\]'

" Byte arrays
syn match chipByteArray 'b\[[^\]]*\]'

" Numbers
syn match chipNumber '\<\d\+\>'
syn match chipFloat '\<\d\+\.\d\+\>'

" Operators
syn match chipOperator '[-+*/%^=!<>]'
syn match chipOperator '=='
syn match chipOperator '!='
syn match chipOperator '<='
syn match chipOperator '>='
syn match chipOperator '<<'
syn match chipOperator '>>'

" Delimiters
syn match chipDelimiter '[(){}\[\],;]'

" Shebang
syn match chipShebang '^#!.*$'

" Highlighting
hi def link chipKeyword Keyword
hi def link chipOperator Operator
hi def link chipConstant Constant
hi def link chipError Error
hi def link chipSuccess String
hi def link chipBuiltin Function
hi def link chipComment Comment
hi def link chipTodo Todo
hi def link chipString String
hi def link chipStringEscape SpecialChar
hi def link chipByteArray Type
hi def link chipNumber Number
hi def link chipFloat Float
hi def link chipDelimiter Delimiter
hi def link chipShebang PreProc

let b:current_syntax = "chiplang"
