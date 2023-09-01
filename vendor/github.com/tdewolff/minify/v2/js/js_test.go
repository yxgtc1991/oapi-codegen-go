package js

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/buffer"
	"github.com/tdewolff/test"
)

func TestJS(t *testing.T) {
	jsTests := []struct {
		js       string
		expected string
	}{
		{`#!shebang`, `#!shebang`},
		{`/*comment*/a`, `a`},
		{`/*!comment*/a`, `/*!comment*/a`},
		{"//!comment1\n\n//!comment2\na", "//!comment1\n//!comment2\na"},
		{`debugger`, `debugger`},
		{`"use strict"`, `"use strict"`},
		{`1.0`, `1`},
		{`1_2.0_3`, `12.03`},
		{`1000`, `1e3`},
		{`1e10`, `1e10`},
		{`1e-10`, `1e-10`},
		{`5_000`, `5e3`},
		{`0b1001`, `9`},
		{`0b10_01`, `9`},
		{`0o11`, `9`},
		{`0x0D`, `13`},
		{`0x0d`, `13`},
		//{`123456787654321`, `0x704885f926b1`},
		//{`4294967295`, `0xFFFFFFFF`}, // better GZIP
		{`+ +x`, `+ +x`},
		{`- -x`, `- -x`},
		{`- +x`, `-+x`},
		{`+ ++x`, `+ ++x`},
		{`- --x`, `- --x`},
		{`- ++x`, `-++x`},
		{`a + ++b`, `a+ ++b`},
		{`a - --b`, `a- --b`},
		{`a++ +b`, `a+++b`},
		{`a-- -b`, `a---b`},
		{`a - ++b`, `a-++b`},
		{`a-- > b`, `a-- >b`},
		{`(a--) > b`, `a-- >b`},
		{`a-- < b`, `a--<b`},
		{`a < !--b`, `a<! --b`},
		{`a > !--b`, `a>!--b`},
		{`!--b`, `!--b`},
		{`/a/ + b`, `/a/+b`},
		{`/a/ instanceof b`, `/a/ instanceof b`},
		{`[a] instanceof b`, `[a]instanceof b`},
		{`let a = 5;a`, `let a=5;a`},
		{`let a = 5,b;a,b`, `let a=5,b;a,b`},
		{`let a,b = 5;a,b`, `let a,b=5;a,b`},
		{`function a(){}`, `function a(){}`},
		{`function a(b){b}`, `function a(b){b}`},
		{`function a(b, c, ...d){}`, `function a(b,c,...d){}`},
		{`function * a(){}`, `function*a(){}`},
		{`function a(){}; return 5`, `function a(){}return 5`},
		{`x = function (){}`, `x=function(){}`},
		{`x = function a(){}`, `x=function(){}`},
		{`x = function (a){a}`, `x=function(a){a}`},
		{`x = function (a, b, ...c){}`, `x=function(a,b,...c){}`},
		{`x = function (){};y=z`, `x=function(){},y=z`},
		{`return 5`, `return 5`},
		{`return .5`, `return.5`},
		{`return-5`, `return-5`},
		{`break a`, `break a`},
		{`continue a`, `continue a`},
		{`label: b`, `label:b`},
		{`typeof a`, `typeof a`},
		{`new RegExp()`, `new RegExp`},
		{`switch (a) { case b: 5; default: 6}`, `switch(a){case b:5;default:6}`},
		{`switch (a) { case b: {var c;return c}; default: 6}`, `switch(a){case b:{var c;return c}default:6}`},
		{`switch (a) { case b: 5 }while(b);`, `switch(a){case b:5}for(;b;);`},
		{`switch (a) { case "text": 5}`, `switch(a){case"text":5}`},
		{`let a=5;switch(b){case 0:let a=5}`, `let a=5;switch(b){case 0:let a=5}`},
		{`with (a = b) x`, `with(a=b)x`},
		{`with (a = b) {x}`, `with(a=b)x`},
		{`import "path"`, `import"path"`},
		{`import x from "path"`, `import x from"path"`},
		{`import * as b from "path"`, `import*as b from"path"`},
		{`import {a as b} from "path"`, `import{a as b}from"path"`},
		{`import {a as b, c} from "path"`, `import{a as b,c}from"path"`},
		{`import x, * as b from "path"`, `import x,*as b from"path"`},
		{`import x, {a as b, c} from "path"`, `import x,{a as b,c}from"path"`},
		{`export * from "path"`, `export*from"path"`},
		{`export * as ns from "path"`, `export*as ns from"path"`},
		{`export {a as b} from "path"`, `export{a as b}from"path"`},
		{`export {a as b, c} from "path"`, `export{a as b,c}from"path"`},
		{`export {a as b, c}`, `export{a as b,c}`},
		{`export var a = b`, `export var a=b`},
		{`export function a(){}`, `export function a(){}`},
		{`export default a = b`, `export default a=b`},
		{`export default a = b;c=d`, `export default a=b;c=d`},
		{`export default function a(){};c=d`, `export default function(){}c=d`},
		{`!class {}`, `!class{}`},
		{`class a {}`, `class a{}`},
		{`class a extends b {}`, `class a extends b{}`},
		{`class a extends(!b){}`, `class a extends(!b){}`},
		{`class a { f(a) {a} }`, `class a{f(a){a}}`},
		{`class a { f(a) {a}; static g(b) {b} }`, `class a{f(a){a}static g(b){b}}`},
		{`class a { static }`, `class a{static}`},
		{`class a { static b }`, `class a{static b}`},
		{`class a { f(c){c} }`, `class a{f(c){c}}`},
		{`class a { static #d=5 }`, `class a{static#d=5}`},
		{`class a { static { b = this.f(5) } }`, `class a{static{b=this.f(5)}}`},
		{`class a { #a(){} }`, `class a{#a(){}}`},
		{`for (var a = 5; a < 10; a++){a}`, `for(var a=5;a<10;a++)a`},
		{`for (a,b = 5; a < 10; a++){a}`, `for(a,b=5;a<10;a++)a`},
		{`async function f(){for await (var a of b){a}}`, `async function f(){for await(var a of b)a}`},
		{`for (var a in b){a}`, `for(var a in b)a`},
		{`for (a in b){a}`, `for(a in b)a`},
		{`for (var a of b){a}`, `for(var a of b)a`},
		{`for (a of b){a}`, `for(a of b)a`},
		{`for (;;){let a;a}`, `for(;;){let a;a}`},
		{`var a;for(var b;;){let a;a++}a,b`, `for(var a,b;;){let a;a++}a,b`},
		{`var a;for(var b;;){let c = 10;c++}a,b`, `for(var a,b;;){let c=10;c++}a,b`},
		{`while(a < 10){a}`, `for(;a<10;)a`},
		{`while(a < 10){a;b}`, `for(;a<10;)a,b`},
		{`while(a < 10){while(b);c}`, `for(;a<10;){for(;b;);c}`},
		//{`while(a) if (!b) break`, `for(;a&&b;);`},
		{`do {a} while(a < 10)`, `do a;while(a<10)`},
		{`do [a]=5; while(a < 10)`, `do[a]=5;while(a<10)`},
		{`do [a]=5; while(a < 10);return a`, `do[a]=5;while(a<10)return a`},
		{`throw a`, `throw a`},
		{`throw [a]`, `throw[a]`},
		{`try {a} catch {b}`, `try{a}catch{b}`},
		{`try {a} catch(b) {b}`, `try{a}catch(b){b}`},
		{`try {a} catch(b) {b} finally {c}`, `try{a}catch(b){b}finally{c}`},
		{`try {a} finally {c}`, `try{a}finally{c}`},
		{`try {a} catch(b) {c}`, `try{a}catch{c}`},
		{`a=b;c=d`, `a=b,c=d`},

		// strings
		{`""`, `""`},
		{`"\x7"`, `"\x7"`},
		{`"string\'string"`, `"string'string"`},
		{`'string\"string'`, `'string"string'`},
		{`"string\t\f\v\bstring"`, "\"string\t\f\v\bstring\""},
		{`"string\a\c\'string"`, `"stringac'string"`},
		{`"string\∀string"`, `"string∀string"`},
		{`"string\0\uFFFFstring"`, "\"string\\0\uffffstring\""},
		{`"string\x00\x55\x0A\x0D\x22\x27string"`, `"string\x00U\n\r\"'string"`},
		{`"string\000\12\015\042\47\411string"`, "\"string\\0\\n\\r\\\"'!1string\""},
		{`"\x005"`, `"\x005"`},
		{"'string\\n\\rstring'", "`string\n\rstring`"},
		{"'string\\\r\nstring\\\nstring\\\rstring\\\u2028string\\\u2029string'", `"stringstringstringstringstringstring"`},
		{`"\x7H\u877H"`, `"\x7H\u877H"`},
		{`"\u01ac\u01de\u0187\u{0001a0}"`, `"ƬǞƇƠ"`},
		{`"str1ng" + "str2ng"`, `"str1ngstr2ng"`},
		{`"str1ng" + "str2ng" + "str3ng"`, `"str1ngstr2ngstr3ng"`},
		{`"padding" + this`, `"padding"+this`},
		{`"<\/script>"`, `"<\/script>"`},
		{`"</scr"+"ipt>"`, `"<\/script>"`},
		{`"\""`, `'"'`},
		{`'\'""'`, "`'\"\"`"},
		{`"\"\"a'"`, "`\"\"a'`"},
		{`"'" + '"'`, "`'\"`"},
		{`'"' + "'"`, "`\"'`"},
		{"`\\n\\'\\$\\$\\{`", "`\n'$\\${`"},
		{`"a"+"b"+5`, `"ab"+5`},
		{`5+"a"+"b"`, `5+"ab"`},
		{`"a"+"b"+5+"c"+"d"`, `"ab"+5+"cd"`},
		{`"a"+"b"+5+6+"d"`, `"ab"+5+6+"d"`},
		{"`$${foo}`", "`$${foo}`"},

		// rename true, false, undefined, Infinity
		{`x=true`, `x=!0`},
		{`x=false`, `x=!1`},
		{`x=false()`, `x=(!1)()`},
		{`false`, `!1`},
		{`x=undefined`, `x=void 0`},
		{`x=undefined()`, `x=(void 0)()`},
		{`x=undefined.a`, `x=(void 0).a`},
		//{`undefined=5;x=undefined`, `undefined=5;x=undefined`},
		{`x=Infinity`, `x=1/0`},
		{`x=Infinity()`, `x=(1/0)()`},
		{`x=2**Infinity`, `x=2**(1/0)`},
		//{`Infinity=5;x=Infinity`, `Infinity=5;x=Infinity`},
		{`class a extends undefined {}`, `class a extends(void 0){}`},
		{`new true`, `new(!0)`},
		{`function*a(){yield undefined}`, `function*a(){yield}`},
		{`function*a(){yield*undefined}`, `function*a(){yield*void 0}`},

		// if/else statements
		{`if(a){return b}`, `if(a)return b`},
		{`if(a){b = 5;return b}`, `if(a)return b=5,b`},
		{`if(a)`, `a`},
		{`if(a){}`, `a`},
		{`if(a) a`, `a&&a`},
		{`if(a) b`, `a&&b`},
		{`if(a,b) c`, `a,b&&c`},
		{`if(a){}else;`, `a`},
		{`if(a){}else{}`, `a`},
		{`if(a){}else{;}`, `a`},
		{`if(a){}else{b}`, `a||b`},
		{`if(a)a;else b`, `a||b`},
		{`if(a)b;else b`, `a,b`},
		{`if(a){b=c}else if(d){e=f}`, `a?b=c:d&&(e=f)`},
		{`if(a){b=c;y=z}else if(d){e=f}`, `a?(b=c,y=z):d&&(e=f)`},
		{`if(a)while(b){c;d}else e`, `if(a)for(;b;)c,d;else e`},
		{`if(a)while(b){c}else e`, `if(a)for(;b;)c;else e`},
		{`if(a){ if(b) c }`, `a&&b&&c`},
		{`if(a){ if(b) c } else e`, `a?b&&c:e`},
		{`if(a){ if(b) c; else d} else e`, `a?b?c:d:e`},
		{`if(a){ if(b) c; else for(x;y;z){f=g}} else e`, `if(a)if(b)c;else for(x;y;z)f=g;else e`},
		{`if(a){ if(b) c; else {for(x;y;z){f=g}}} else e`, `if(a)if(b)c;else for(x;y;z)f=g;else e`},
		{`if(a)a={b};else e`, `a?a={b}:e`},
		{`if(a) a; else [e]=4`, `a?a:[e]=4`},
		{`if(a){ a = b?c:function(d){d} } else e`, `a?a=b?c:function(d){d}:e`},
		{`if(a)while(b){if(c)d; else e}else f`, `if(a)for(;b;)c?d:e;else f`},
		{`if(a)b=c`, `a&&(b=c)`},
		{`if(!a)b=c`, `a||(b=c)`},
		{`if(a||d)b=c`, `(a||d)&&(b=c)`},
		{`if(a);else b=c`, `a||(b=c)`},
		{`if(!a);else b=c`, `a&&(b=c)`},
		{`if(a)b=c;else e`, `a?b=c:e`},
		{`if(a)b=c,f;else e`, `a?(b=c,f):e`},
		{`if(a){b=c}else{if(d){e=f}else{g=h}}`, `a?b=c:d?e=f:g=h`},
		{`if(a){b=c}else if(d){e=f}else if(g){h=i}`, `a?b=c:d?e=f:g&&(h=i)`},
		{`if(a){if(b)c;else d}else{e}`, `a?b?c:d:e`},
		{`if(a){if(b)c;else d}else{d}`, `a&&b?c:d`},
		{`if(a){if(b)c;else false}else{d}`, `a?!!b&&c:d`},
		{`if(a){if(b)c;else d}else{false}`, `!!a&&(b?c:d)`},
		{`if(a){if(b)c;else false}else{false}`, `!!a&&!!b&&c`},
		{`if(a)return a;else return b`, `return a||b`},
		{`if(a)return a;else a++`, `if(a)return a;a++`},
		{`if(a)return b;else a++`, `if(a)return b;a++`},
		{`if(a)throw b;else a++`, `if(a)throw b;a++`},
		{`if(a)break;else a++`, `if(a)break;a++`},
		{`if(a)return;else return`, `a`},
		{`if(a)throw a;else throw b`, `throw a||b`},
		{`if(a)return a;else a=b`, `if(a)return a;a=b`},
		{`if(a){a++;return a}else a=b`, `if(a)return a++,a;a=b`},
		{`if(a){a++;return a}else if(b)a=b`, `if(a)return a++,a;b&&(a=b)`},
		{`if(a){a++;return}else a=b`, `if(a){a++;return}a=b`},
		//{`if(a){a++;return}else return`, `if(a){a++}return`}, // TODO
		//{`if(a){a++;return}return`, `if(a){a++}return`}, // TODO
		{`if(a){return}else {a=b;while(c){}}`, `if(a)return;for(a=b;c;);`},
		{`if(a){a++;return a}else return`, `if(a)return a++,a`},
		{`if(a){a++;return a}return`, `if(a)return a++,a`},
		{`if(a){return a}return b`, `return a||b`},
		{`if(a);else return a;return b`, `return a&&b`},
		{`if(a){return a}b=c;return b`, `return a||(b=c,b)`},
		{`if(a){return}b=c;return b`, `if(a)return;return b=c,b`},
		{`if(a){return a}b=c;return`, `if(a)return a;b=c`},
		{`if(a){return}return`, `a`},
		{`if(a);else{return}return`, `a`},
		{`if(a){throw a}b=c;throw b`, `throw a||(b=c,b)`},
		{`if(a);else{throw a}b=c;throw b`, `throw a&&(b=c,b)`},
		{`if(a)a++;else b;if(b)b++;else c`, `a?a++:b,b?b++:c`},
		{`if(a){while(b);}`, `if(a)for(;b;);`},
		{`if(a){while(b);c}`, `if(a){for(;b;);c}`},
		{`if(a){if(b){while(c);}}`, `if(a&&b)for(;c;);`},
		{`if(a){}else{while(b);}`, `if(a);else for(;b;);`},
		{`if(a){return b}else{while(c);}`, `if(a)return b;for(;c;);`},
		{`if(a){return b}else{while(c);d}`, `if(a)return b;for(;c;);d`},
		{`if(!a){while(b);c}`, `if(!a){for(;b;);c}`},
		{`while(a){if(b)continue;if(c)continue;else d}`, `for(;a;){if(b)continue;if(c)continue;d}`},
		{`while(a)if(b)continue;else c`, `for(;a;){if(b)continue;c}`},
		{`while(a)if(b)return c;else return d`, `for(;a;)return b?c:d`},
		{`while(a){if(b)continue;else c}`, `for(;a;){if(b)continue;c}`},
		{`if(a){while(b)if(c)5}else{6}`, `if(a)for(;b;)c&&5;else 6`},
		{`if(a){for(;;)if(b)break}else c`, `if(a){for(;;)if(b)break}else c`},
		{`if(a){for(d in e)if(b)break}else c`, `if(a){for(d in e)if(b)break}else c`},
		{`if(a){for(d of e)if(b)break}else c`, `if(a){for(d of e)if(b)break}else c`},
		{`if(a){for(d of e)if(f)g()}else c`, `if(a)for(d of e)f&&g();else c`},
		{`if(a){d:if(b)break}else c`, `if(a){d:if(b)break}else c`},
		{`if(a){with(d)if(b)break}else c`, `if(a){with(d)if(b)break}else c`},
		{`if(a)return b;if(c)return d;return e`, `return a?b:c?d:e`},
		{`if(a,b)b`, `a,b&&b`},
		{`if(a,b)b;else d`, `a,b||d`},
		{`if(a=b)a;else b`, `(a=b)||b`},
		{`if(!a&&!b){return true}else if(!a||!b){return false}return c&&d`, `return!a&&!b||!!a&&!!b&&c&&d`},
		{`if(!a){if(b){throw c}else{return c}}else{return a}`, `if(a)return a;if(b)throw c;return c`},
		{`if(!a){return y}else if(b){if(c){return x}}return z`, `return a?b&&c?x:z:y`},
		{`if(a)b:{if(c)break b}else if(d)e()`, `if(a){b:if(c)break b}else d&&e()`},

		// var declarations
		{`var a;var b;a,b`, `var a,b;a,b`},
		{`const a=1;const b=2;a,b`, `const a=1,b=2;a,b`},
		{`let a=1;let b=2;a,b`, `let a=1,b=2;a,b`},
		{`var a;if(a)var b;else b`, `var a,b;a||b`},
		{`var a;if(a)var b=5;b`, `if(a)var a,b=5;b`}, // TODO: or should we try to take var decls out of statements that will be converted to expressions?
		{`var a;for(var b=0;b;b++);a`, `for(var a,b=0;b;b++);a`},
		{`var a=1;for(var b=0;b;b++);a`, `for(var a=1,b=0;b;b++);a`},
		{`var a=1;for(var a;a;a++);`, `for(var a=1;a;a++);`},
		{`var a;for(var a=1;a;a++);`, `for(var a=1;a;a++);`},
		{`var [,,a,,]=b`, `var[,,a]=b`},
		{`var [,,a,,...c]=b`, `var[,,a,,...c]=b`},
		{`const a=3;for(const b=0;b;b++);a`, `const a=3;for(const b=0;b;b++);a`},
		{`var a;for(let b=0;b;b++);a`, `var a;for(let b=0;b;b++);a`},
		{`var [a,]=[b,]`, `var[a]=[b]`},
		{`var [a,b=5,...c]=[d,e,...f]`, `var[a,b=5,...c]=[d,e,...f]`},
		{`var [a,,]=[b,,]`, `var[a]=[b,,]`},
		{`var {a,}=b`, `var{a}=b`},
		{`var {a:a}=b`, `var{a}=b`},
		{`var {a,b=5,...c}={d,e=7,...f}`, `var{a,b=5,...c}={d,e=7,...f}`},
		{`var {[a+b]: c}=d`, `var{[a+b]:c}=d`},
		{`for(var [a] in b);`, `for(var[a]in b);`},
		{`for(var {a} of b);`, `for(var{a}of b);`},
		{`for(var a in b);var c`, `for(a in b);var a,c`},
		{`for(var a in b);var c=6,d=7`, `for(a in b);var a,c=6,d=7`},
		{`for(var a=5,c=6;;);`, `for(var a=5,c=6;;);`},
		{`while(a);var b;var c`, `for(var b,c;a;);`},
		{`while(a){d()}var b;var c`, `for(var b,c;a;)d()`},
		{`var [a,b=5,,...c]=[d,e,...f];var z;z`, `var[a,b=5,,...c]=[d,e,...f],z;z`},
		{`var {a,b=5,[5+8]:c,...d}={d,e,...f};var z;z`, `var{a,b=5,[5+8]:c,...d}={d,e,...f},z;z`},
		{`var a=5;var b=6;a,b`, `var a=5,b=6;a,b`},
		{`var a;var b=6;a=7;b`, `var b=6,a=7;b`}, // swap declaration order to maintain definition order
		{`var a=5;var b=6;a=7,b`, `var a=5,b=6,a=7;b`},
		{`var a;var b=6;a,b,z=7`, `var a,b=6;a,b,z=7`},
		{`for(var a=6,b=7;;);var c=8;a,b,c`, `for(var c,a=6,b=7;;);c=8,a,b,c`},
		{`for(var c;b;){let a=8;a};var a;a`, `for(var a,c;b;){let a=8;a}a`},
		{`for(;b;){let a=8;a};var a;var b;a`, `for(var a,b;b;){let a=8;a}a`},
		{`var a=1,b=2;while(c);var d=3,e=4;a,b,d,e`, `for(var d,e,a=1,b=2;c;);d=3,e=4,a,b,d,e`},
		{`var z;var [a,b=5,,...c]=[d,e,...f];z`, `var[a,b=5,,...c]=[d,e,...f],z;z`},
		{`var z;var {a,b=5,[5+8]:c,...d}={d,e,...f};z`, `var{a,b=5,[5+8]:c,...d}={d,e,...f},z;z`},
		{`var z;z;var [a,b=5,,...c]=[d,e,...f];a`, `z;var[a,b=5,,...c]=[d,e,...f],z;a`},
		// TODO
		//{`var z;z;var {a,b=5,[5+8]:c,...d}={e,f,...g};a`, `var z,a;z,{a}={e,f,...g},a`},
		//{`var z;z;var {a,b=5,[5+8]:c,...d}={e,f,...g};d`, `var z,a,b,c,d;z,{a,b,[5+8]:c,...d}={e,f,...g},d`},
		//{`var {a,b=5,[5+8]:c,d:e}=z;b`, `var{b=5}=z;b`},
		//{`var {a,b=5,[5+8]:c,d:e,...f}=z;b`, `var{b=5}=z;b`},
		{`var {a,b=5,[5+8]:c,d:e,...f}=z;f`, `var{a,b=5,[5+8]:c,d:e,...f}=z;f`},
		{`var a;var {}=b;`, `var{}=b,a`},
		{`"use strict";var a;var b;b=5`, `"use strict";var a,b=5`},
		{`"use strict";z+=6;var a;var b;b=5`, `"use strict";z+=6;var a,b=5`},
		{`!function(){"use strict";return a}`, `!function(){"use strict";return a}`},
		{`var a;var{b}=c;`, `var{b}=c,a`},
		{`var a;var[b]=c;`, `var[b]=c,a`},
		{`var a;f();var b=c;`, `f();var a,b=c`},
		{`var{a}=x;f();var{b}=c;`, `var{a}=x;f();var{b}=c`},
		{`var{a}=x;f();var d,{b}=c;`, `var{a}=x;f();var{b}=c,d`},
		{`var{a}=x;f();var[b]=c;`, `var{a}=x,b;f(),[b]=c`},
		{`var{a}=x;f();var[b,d]=c;`, `var{a}=x;f();var[b,d]=c`},
		{`var{a}=x;f();var[bd]=c;`, `var{a}=x,bd;f(),[bd]=c`},
		{`var{a}=x;f();var bc=e;`, `var{a}=x,bc;f(),bc=e`},
		{`var{a}=x;f();var bcd=e;`, `var{a}=x,bcd;f(),bcd=e`},
		{`var{a}=x;f();var b,d;`, `var{a}=x,b,d;f()`},
		{`var{a}=x;f();var b,d=e;`, `var{a}=x,b,d;f(),d=e`},
		{`var{a}=x;f();var b=c,d=e;`, `var{a}=x,b,d;f(),b=c,d=e`},
		{`var{a}=x;f();var[b]=c,d=e;`, `var{a}=x;f();var[b]=c,d=e`},
		// {`var{a}=x;f();var{b}=y`, `var{a}=x,b;f(),{b}=y`}, // we can't know that {b} doesn't require parentheses
		{`var a=0;a=1`, `var a=0,a=1`},
		{`var a,b;a=b`, `var b,a=b`},
		{`var a,b=c;a=b`, `var b=c,a=b`},
		{`var a,b=c;b=a`, `var a,b=c,b=a`},
		{`var{a}=f;var b=c,d=e;`, `var{a}=f,b=c,d=e`},
		{`var a,b;a=1,b=2,c=3`, `var a=1,b=2;c=3`},
		{`var a=[];var b={};var c=d,e=f`, `var a=[],b={},c=d,e=f`},
		{`var a=[];var b={};var c=d,e=f`, `var a=[],b={},c=d,e=f`},
		{`var a=[];var b;var c,e=f`, `var b,c,a=[],e=f`},
		{`var a=[];f();var b;f();var c;f();var e=f`, `var b,c,e,a=[];f(),f(),f(),e=f`},
		{`var {...a}=c;for(var {...b}=d;b;b++);`, `for(var{...a}=c,{...b}=d;b;b++);`},
		{`var o=8,p=9,x=0,y=1;x=x+2;y=y+3;var b=1,c=2,d=3`, `var o=8,p=9,x=0,y=1,x=x+2,y=y+3,b=1,c=2,d=3`},
		{`var result=[];for(var a=0,array=d;a<array.length;a++){var v=array[a]}`, `for(var v,result=[],a=0,array=d;a<array.length;a++)v=array[a]`},
		//{`var name=function name(){name()}`, `var name=function(){name()}`}, // TODO
		{`var a=0,b=a;var c=1,d=2,e=3`, `var a=0,b=a,c=1,d=2,e=3`},

		// TODO: test for variables renaming (first rename, then merge vars)

		// function and method declarations
		{`function g(){return}`, `function g(){}`},
		{`function g(){return undefined}`, `function g(){}`},
		{`function g(){return void 0}`, `function g(){}`},
		{`function g(){return a++,void 0}`, `function g(){a++}`},
		{`for (var a of b){continue}`, `for(var a of b);`},
		{`for (var a of b){continue LABEL}`, `for(var a of b)continue LABEL`},
		{`for (var a of b){break}`, `for(var a of b)break`},
		{`class a{static g(){}}`, `class a{static g(){}}`},
		{`class a{static [1](){}}`, `class a{static[1](){}}`},
		{`class a{static*g(){}}`, `class a{static*g(){}}`},
		{`class a{static*[1](){}}`, `class a{static*[1](){}}`},
		{`class a{get g(){}}`, `class a{get g(){}}`},
		{`class a{get [1](){}}`, `class a{get[1](){}}`},
		{`class a{set g(){}}`, `class a{set g(){}}`},
		{`class a{set [1](){}}`, `class a{set[1](){}}`},
		{`class a{static async g(){}}`, `class a{static async g(){}}`},
		{`class a{static async [1](){}}`, `class a{static async[1](){}}`},
		{`class a{static async*g(){}}`, `class a{static async*g(){}}`},
		{`class a{static async*[1](){}}`, `class a{static async*[1](){}}`},
		{`class a{"f"(){}}`, `class a{f(){}}`},
		{`class a{f(){};g(){}}`, `class a{f(){}g(){}}`},
		{`class a{one;#two = 2;f(){}}`, `class a{one;#two=2;f(){}}`},
		//{`function g(){a()} function g(){b()}`, `function g(){b()}`},

		// dead code
		//{`return;a`, `return`},
		//{`break;a`, `break`},
		//{`if(a){return;a=5;b=6}`, `if(a)return`},
		//{`if(a){throw a;a=5}`, `if(a)throw a`},
		//{`if(a){break;a=5}`, `if(a)break`},
		//{`if(a){continue;a=5}`, `if(a)continue`},
		//{`if(a){return a;a=5}return b`, `return a||b`},
		//{`if(a){throw a;a=5}throw b`, `throw a||b`},
		//{`if(a){return;var b}return`, `a;var b`},
		//{`if(a){return;function b(){}}`, `a;var b`},
		//{`for (var a of b){continue;var c}`, `for(var a of b){}var c`},
		//{`if(false)a++;else b`, `b`},
		//{`if(false){var a}`, `var a`},
		//{`if(false){var a;a++}else b`, `var a;b`},
		//{`if(false){function a(c){return d};a++}else b`, `var a;b`},
		//{`if(!1)a++;else b`, `b`},
		//{`if(null)a++;else b`, `b`},
		//{`var a;if(false)var b`, `var a,b`},
		//{`var a;if(false)var b=5`, `var a,b`},
		//{`var a;if(false){const b}`, `var a`},
		//{`var a;if(false){function b(){}}`, `var a;if(!1)function b(){}`},
		//{`function f(){if(a){a=5;return}a=6;return a}`, `function f(){if(!a){a=6;return a}a=5}`},
		//{`function g(){return;var a;a=b}`, `function g(){var a;}`},
		//{`function g(){return 5;function f(){}}`, `function g(){return 5;function f(){}}`},
		//{`function g(){if(a)return a;else return b;var c;c=d}`, `function g(){var c;return a||b}`},
		//{`()=>a()`, ``},

		// arrow functions
		{`() => {}`, `()=>{}`},
		{`(a) => {a}`, `a=>{a}`},
		{`(...a) => {}`, `(...a)=>{}`},
		{`(a=0) => {a}`, `(a=0)=>{a}`},
		{`(a,b) => {a,b}`, `(a,b)=>{a,b}`},
		{`a => {a++}`, `a=>{a++}`},
		{`x=(a) => {a}`, `x=a=>{a}`},
		{`x=() => {return}`, `x=()=>{}`},
		{`x=(a) => {return a}`, `x=a=>a`},
		{`x=(a) => {a++;return a}`, `x=a=>(a++,a)`},
		{`x=(a) => {while(b);return}`, `x=a=>{for(;b;);}`},
		{`x=(a) => {while(b);return a}`, `x=a=>{for(;b;);return a}`},
		{`x=(a) => {a++}`, `x=a=>{a++}`},
		{`x=(a) => {a++}`, `x=a=>{a++}`},
		{`x=(a,b) => a+b`, `x=(a,b)=>a+b`},
		{`async a => await b`, `async a=>await b`},
		{`([])=>5`, `([])=>5`},
		{`({})=>5`, `({})=>5`},

		// remove groups
		{`a=(b+c)+d`, `a=b+c+d`},
		{`a=b+(c+d)`, `a=b+(c+d)`},
		{`a=b*(c+d)`, `a=b*(c+d)`},
		{`a=(b*c)+d`, `a=b*c+d`},
		{`a=(b.c)++`, `a=b.c++`},
		{`a=(b++).c`, `a=(b++).c`},
		{`a=!(b++)`, `a=!b++`},
		{`a=(b+c)(d)`, `a=(b+c)(d)`},
		{`a=b**(c**d)`, `a=b**c**d`},
		{`a=(b**c)**d`, `a=(b**c)**d`},
		{`a=false**2`, `a=(!1)**2`},
		{`a=(++b)**2`, `a=++b**2`},
		{`a=(a||b)&&c`, `a=(a||b)&&c`},
		{`a=a||(b&&c)`, `a=a||b&&c`},
		{`a=(a&&b)||c`, `a=a&&b||c`},
		{`a=a&&(b||c)`, `a=a&&(b||c)`},
		{`a=a&&(b&&c)`, `a=a&&b&&c`},
		{`a=c&&(a??b)`, `a=c&&(a??b)`},
		{`a=(a||b)??(c||d)`, `a=(a||b)??(c||d)`},
		{`a=(a&&b)??(c&&d)`, `a=(a&&b)??(c&&d)`},
		{`a=(a??b)??(c??d)`, `a=a??b??c??d`},
		{`a=(a||b)||(c||d)`, `a=a||b||c||d`},
		{`a=!(!b)`, `a=!!b`},
		{`a=(b())`, `a=b()`},
		{`a=(b)?.(c,d)`, `a=b?.(c,d)`},
		{`a=(b,c)?.(d)`, `a=(b,c)?.(d)`},
		{`a=(b?c:e)?.(d)`, `a=(b?c:e)?.(d)`},
		{`a=b?c:c`, `a=(b,c)`},
		{`a=b?b:c=f`, `a=b?b:c=f`}, // don't write as a=b||(c=f)
		{`a=b||(c=f)`, `a=b||(c=f)`},
		{`a=(-5)**3`, `a=(-5)**3`},
		{`a=5**(-3)`, `a=5**-3`},
		{`a=(-(+5))**3`, `a=(-+5)**3`}, // could remove +
		{`a=(b,c)+3`, `a=(b,c)+3`},
		{`(a,b)&&c`, `a,b&&c`},
		{`function*x(){a=(yield b)}`, `function*x(){a=yield b}`},
		{`function*x(){a=yield (yield b)}`, `function*x(){a=yield yield b}`},
		{`if((a))while((b));`, `if(a)for(;b;);`},
		{`({a}=5)`, `({a}=5)`},
		{`({a:a}=5)`, `({a}=5)`},
		{`({a:"a"}=5)`, `({a:"a"}=5)`},
		{`(function(){})`, `!function(){}`},
		{`(function(){}())`, `(function(){})()`},
		{`(function(){})()`, `(function(){})()`},
		{`(function(){})();x=5;f=6`, `(function(){})(),x=5,f=6`},
		{`(async function(){})`, `!async function(){}`},
		{`(class a{})`, `!class a{}`},
		{`(let [a])`, `!let[a]`},
		{`x=(function(){})`, `x=function(){}`},
		{`x=(function(){}())`, `x=function(){}()`},
		{`x=(function(){})()`, `x=function(){}()`},
		{`x=(function(){}).a`, `x=function(){}.a`},
		{`function g(){await(x+y)}`, `function g(){await(x+y)}`},
		{`async function g(){await(x+y)}`, `async function g(){await(x+y)}`},
		{`function g(){await(fun()())}`, `function g(){await(fun()())}`},
		{`async function g(){await(fun()())}`, `async function g(){await fun()()}`},
		{`function g(){await((fun())())}`, `function g(){await(fun()())}`},
		{`a=1+"2"+(3+4)`, `a=1+"2"+(3+4)`},
		{`(-1)()`, `(-1)()`},
		{`(-1)(-2)`, `(-1)(-2)`},
		{`(+new Date).toString(32)`, `(+new Date).toString(32)`},
		{`(2).toFixed(0)`, `2..toFixed(0)`},
		{`(0.2).toFixed(0)`, `.2.toFixed(0)`},
		{`(2e-8).toFixed(0)`, `2e-8.toFixed(0)`},
		{`(-2).toFixed(0)`, `(-2).toFixed(0)`},
		{`(a)=>((b)=>c)`, `a=>b=>c`},
		{`function f(a=(3+2)){a}`, `function f(a=3+2){a}`},
		{`function*a(){yield a.b}`, `function*a(){yield a.b}`},
		{`function*a(){(yield a).b}`, `function*a(){(yield a).b}`},
		{`function*a(){yield a["-"]}`, `function*a(){yield a["-"]}`},
		{`function*a(){(yield a)["-"]}`, `function*a(){(yield a)["-"]}`},
		{`new(a(b))`, `new(a(b))`},
		{`new(new a)`, `new new a`},
		{`new new a()()`, `new new a`},
		{`new(new a(b))`, `new new a(b)`},
		{`new(a(b))(c)`, `new(a(b))(c)`},
		{`new(a(b).c)(d)`, `new(a(b).c)(d)`},
		{`new(a(b)[5])(d)`, `new(a(b)[5])(d)`},
		{"new(a(b)`tmpl`)(d)", "new(a(b)`tmpl`)(d)"},
		{`new(a(b)).c(d)`, `new(a(b)).c(d)`},
		{`new(a(b))[5](d)`, `new(a(b))[5](d)`},
		{"new(a(b))`tmpl`(d)", "new(a(b))`tmpl`(d)"},
		{`new a().b(c)`, `(new a).b(c)`},
		{`(new a).b(c)`, `(new a).b(c)`},
		{`(new a.b).c(d)`, `(new a.b).c(d)`},
		{`(new a(b)).c(d)`, `new a(b).c(d)`},
		{`(new a().b).c(d)`, `(new a).b.c(d)`},
		{`new a()`, `new a`},
		{`new a()()`, `(new a)()`},
		{`new(a.b)instanceof c`, `new a.b instanceof c`},
		{`new(a[b])instanceof c`, `new a[b]instanceof c`},
		{"new(a`tmpl`)instanceof c", "new a`tmpl`instanceof c"},
		{`(a()).b(c)`, `a().b(c)`},
		{`(a()[5]).b(c)`, `a()[5].b(c)`},
		{"(a()`tmpl`).b(c)", "a()`tmpl`.b(c)"},
		{`(a?.b).c(d)`, `a?.b.c(d)`},
		{`(a?.(c)).d(e)`, `a?.(c).d(e)`},
		{`class a extends (new b){}`, `class a extends new b{}`},
		{`(new.target)`, `new.target`},
		{`(import.meta)`, `(import.meta)`},
		{"(`tmpl`)", "`tmpl`"},
		{"(a`tmpl`)", "a`tmpl`"},
		{"a=-(b=5)", "a=-(b=5)"},
		{"f({},(a=5,b))", "f({},(a=5,b))"},
		{"for(var a=(b in c);;);", "for(var a=(b in c);;);"},
		{`(1,2,a=3)&&b`, `(1,2,a=3)&&b`},
		{`(1,2,a||3)&&b`, `(1,2,a||3)&&b`},
		{`(1,2,a??3)&&b`, `(1,2,a??3)&&b`},
		{`(1,2,a&&3)&&b`, `1,2,a&&3&&b`},
		{`(1,2,a|3)&&b`, `1,2,a|3&&b`},
		{`(a,b)?c:b`, `a,b&&c`},
		{`(a,b)?c:d`, `a,b?c:d`},
		{`f(...a,...b)`, `f(...a,...b)`},

		// expressions
		//{`a=a+5`, `a+=5`},
		//{`a=5+a`, `a+=5`},
		{`a?true:false`, `!!a`},
		{`a==b?true:false`, `a==b`},
		{`!a?true:false`, `!a`},
		{`a?false:true`, `!a`},
		{`!a?false:true`, `!!a`},
		{`a?!0:!1`, `!!a`},
		{`a?0:1`, `a?0:1`},
		{`!!a?0:1`, `!!a?0:1`},
		{`a&&b?!1:!0`, `!a||!b`},
		{`a&&b?!0:!1`, `!!(a&&b)`},
		{`a?true:5`, `!!a||5`},
		{`a?5:false`, `!!a&&5`},
		{`!a?true:5`, `!a||5`},
		{`!a?5:false`, `!a&&5`},
		{`a==b?true:5`, `a==b||5`},
		{`a!=b?true:5`, `a!=b||5`},
		{`a==b?false:5`, `a!=b&&5`},
		{`a!=b?false:5`, `a==b&&5`},
		{`a===b?false:5`, `a!==b&&5`},
		{`a!==b?false:5`, `a===b&&5`},
		{`a==b?5:true`, `a!=b||5`},
		{`a==b?5:false`, `a==b&&5`},
		{`a<b?5:true`, `!(a<b)||5`},
		{`!(a<b)?5:true`, `a<b||5`},
		{`!true?5:true`, `!0`},
		{`true?a:b`, `a`},
		{`false?a:b`, `b`},
		{`!false?a:b`, `a`},
		{`!!false?a:b`, `b`},
		{`!!!false?a:b`, `a`},
		{`undefined?a:b`, `b`},
		{`NaN?a:b`, `b`},
		{`1?a:b`, `a`},
		{`0.00e100?a:b`, `b`},
		{`0x00?a:b`, `b`},
		{`0B00?a:b`, `b`},
		{`0o00?a:b`, `b`},
		{`0n?a:b`, `b`},
		{`(0n?a:b)()`, `b()`},
		{`!0`, `!0`},
		{`!42`, `!1`},
		{`!"str"`, `!1`},
		{`!/regexp/`, `!1`},
		{`typeof a==="object"`, `typeof a=="object"`},
		{`typeof a!=="object"`, `typeof a!="object"`},
		{`"object"===typeof a`, `"object"==typeof a`},
		{`"object"!==typeof a`, `"object"!=typeof a`},
		{`typeof a===b`, `typeof a===b`},
		//{`typeof a==="undefined"`, `typeof a<"u"`}, // only for >ES2020 and not IE
		{`a!=null?a:b`, `a??b`},
		{`a==null?b:a`, `a??b`},
		{`a!=undefined?a:b`, `a??b`},
		{`a==undefined?b:a`, `a??b`},
		{`a==null?true:a`, `a??!0`},
		{`null==a?true:a`, `a??!0`},
		{`a!=null?a:true`, `a??!0`},
		{`a==undefined?true:a`, `a??!0`},
		{`a!=undefined?a:true`, `a??!0`},
		{`a?a:b`, `a||b`},
		{`a?b:a`, `a&&b`},
		{`a&&=b`, `a&&=b`},
		{`a||=b`, `a||=b`},
		{`a??=b`, `a??=b`},
		{`a==false`, `a==!1`},
		{`a===false`, `a===!1`},
		{`!(a||b)`, `!a&&!b`},
		{`!(a&&b)`, `!a||!b`},
		{`!(a&&b)&&c`, `!(a&&b)&&c`},
		{`c&&!(a&&b===5)`, `c&&!(a&&b===5)`},
		{`c&&!(!a&&b!==5)`, `c&&!(!a&&b!==5)`},
		{`c&&!(a==3&&b!==5)`, `c&&(a!=3||b===5)`},
		{`!(a>=0&&a<=1||a>=2&&a<=3)`, `!(a>=0&&a<=1||a>=2&&a<=3)`},
		{`!(0<1||1<2)`, `!(0<1||1<2)`},
		{`!(0<1&&1<2)`, `!(0<1&&1<2)`},
		{`!(a&&b||c&&d)`, `!(a&&b||c&&d)`},
		{`!((a||b)&&(c||d))`, `!a&&!b||!c&&!d`},
		{`a===false||b===true?false:true`, `a!==!1&&b!==!0`},
		//{`!(!(a>=0||a<=1)&&!(a>=2||a<=3))`, `!!(a>=0||a<=1||a>=2||a<=3)`}, // TODO
		{`!!(a===null||a===undefined)`, `a==null`},
		{`a!==null&&a!==undefined`, `a!=null`},
		{`a===null||a===void 0`, `a==null`},
		{`!!(a===b||c===d)`, `a===b||c===d`},
		{`!(a!==null)`, `a===null`},
		{`a==void 0`, `a==null`},
		{`a?b(c):b(d)`, `b(a?c:d)`},
		//{`if(a!==null&&a!==undefined)a.b()`, `a?.b()`},  // returns undefined instead of false
		{`(a===null||a===undefined)?undefined:a()`, `a?.()`},
		{`(a===null||a===undefined)?undefined:a[0]`, `a?.[0]`},
		{"(a===null||a===undefined)?undefined:a`tpl`", "a?.`tpl`"},
		{`(a===null||a===undefined)?undefined:a.b`, `a?.b`},
		{`(a===null||a===undefined)?undefined:a.b()`, `a?.b()`},
		{`(a===null||a===undefined)?undefined:a.b[0]`, `a?.b[0]`},
		{"(a===null||a===undefined)?undefined:a.b`tpl`", "a?.b`tpl`"},
		{`(a===null||a===undefined)?undefined:a.#b`, `a?.#b`},
		{`(((a===null)||(a===undefined)))?undefined:a()`, `a?.()`},
		//{`(a.b===null||a.b===undefined)?undefined:a.b()`, `a.b?.()`},

		// other
		{`async function g(){await x+y}`, `async function g(){await x+y}`},
		{`a={"property": val1, "2": val2, "3name": val3};`, `a={property:val1,2:val2,"3name":val3}`},
		{`a={"key'\"": v,};`, `a={"key'\"":v}`},
		{`() => { const v=6; x={v} }`, `()=>{const v=6;x={v}}`},
		{`a=obj["if"]`, `a=obj.if`},
		{`a=obj["2"]`, `a=obj[2]`},
		{`a=obj["3name"]`, `a=obj["3name"]`},
		{"a=b`tmpl${a?b:b}tmpl`", "a=b`tmpl${a,b}tmpl`"},
		{`a=b?.[c]`, `a=b?.[c]`},
		{`a=b.#c`, `a=b.#c`},
		{`a=b().#c`, `a=b().#c`},
		{`a=b?.#c`, `a=b?.#c`},
		{`a={b(c){c}}`, `a={b(c){c}}`},
		{`a(b,...c)`, `a(b,...c)`},
		{`let a="string";a`, `let a="string";a`},
		{`f((a,b)||d)`, `f((a,b)||d)`},

		// merge expressions
		{`b=5;return a+b`, `return b=5,a+b`},
		{`b=5;throw a+b`, `throw b=5,a+b`},
		{`a();b();return c()`, `return a(),b(),c()`},
		{`a();b();throw c()`, `throw a(),b(),c()`},
		{`a=b;if(a){return a}else return b`, `return a=b,a||b`},
		{`a=5;if(b)while(c);`, `if(a=5,b)for(;c;);`},
		{`a=5;while(b)c()`, `for(a=5;b;)c()`},
		{`a=5;while(b){c()}`, `for(a=5;b;)c()`},
		{`a=5;for(;b;)c()`, `for(a=5;b;)c()`},
		{`a=5;for(b=4;b;)c()`, `a=5;for(b=4;b;)c()`},
		//{`a in 5;for(;b;)c()`, `a in 5;for(;b;)c()`}, // TODO
		{`a in 5;for(b=4;b;)c()`, `a in 5;for(b=4;b;)c()`},
		{`var a=5;for(;a;)c()`, `for(var a=5;a;)c()`},
		{`let a=5;for(;a;)c()`, `let a=5;for(;a;)c()`},
		{`var a=b in c;for(;a;)c()`, `for(var a=(b in c);a;)c()`},
		{`var a=5;for(var a=6,b;b;)c()`, `for(var b,a=5,a=6;b;)c()`},
		{`var a=5;for(var a,b;b;)c()`, `for(var b,a=5;b;)c()`},
		//{`var a=5;for(var b=6,c=7;;);`, `for(var a=5,b=6,c=7;;);`}, // TODO
		{`var a=5;while(a)c()`, `for(var a=5;a;)c()`},
		{`var a=5;while(a){c()}`, `for(var a=5;a;)c()`},
		{`let a=5;while(a)c()`, `let a=5;for(;a;)c()`},
		//{`var a;for(a=5;b;)c()`, `for(var a=5;b;)c()`}, // TODO
		{`a=5;for(var b=4;b;)c()`, `a=5;for(var b=4;b;)c()`},
		{`a=5;switch(b=4){}`, `switch(a=5,b=4){}`},
		{`a=5;with(b=4){}`, `with(a=5,b=4);`},
		{`(function(){})();(function(){})()`, `(function(){})(),function(){}()`},

		// collapse functions
		//{`var a=function(){return 5}`, `var a=()=>5`},
		//{`var a=async function(b){b=6;return 5}`, `var a=async b=>(b=6,5)`},
		//{`(function(){return 5})()`, `(()=>5)()`},
		//{`class c{a(){return 5}}`, `class c{a:()=>5}`},
		//{`export default{a(){return 5}}`, `export default{a:()=>5}`},
		//{`var v={async [[1]](a){return a}}`, `var v={[[1]]:async a=>a}`},
		//{`var a={b:()=>c=5}`, `var a={b(){c=5}}`},
		//{`var a={b:function(){c=5}}`, `var a={b(){c=5}}`},
		//{`var a={b:async function(){c=5}}`, `var a={async b(){c=5}}`},
		//{`var a={b:function*(){c=5}}`, `var a={*b(){c=5}}`},
		//{`a=function(){return 5}()`, `a=5`}, // TODO
		//{`a=function(){if(b){return 3}return 5}()`, `a=b?3:5`}, // TODO

		// collapse variables
		//{`{let a}`, ``}, // TODO
		//{`a=5;b=a;c=b+4`, `c=5+4`}, // TODO
		//{`const a=6;f(a)`, `f(6)`},             // TODO: inline single-use variables that are literals
		//{`let a="string";f(a)`, `f("string")`}, // TODO: inline single-use variables that are literals
		//{`{let a="string"}a`, `a`},
		//{`!function(){var a}`, `!function(){}`}, // TODO
		//{`'a b c'.split(' ')`, `['a','b','c']`}, // TODO?

		// regexps
		{`/\\\/\[/`, `/\\\/\[/`},
		{`/[\\\]]/`, `/[\\\]]/`},
		{`/[\[]/`, `/[[]/`},
		{`/\.\cA\x10\u0010\p{x}\P{x}\0\f\v\n\r\t\S\s\W\w\D\d\b\B\k/`, `/\.\cA\x10\u0010\p{x}\P{x}\0\f\v\n\r\t\S\s\W\w\D\d\b\B\k/`},
		{`/\^\|\(\)\1\*\+\?\{\$/`, `/\^\|\(\)\1\*\+\?\{\$/`},
		{`/[^\-]/`, `/[^-]/`},
		{`/[^\-\-]/`, `/[^--]/`},
		{`/[\^\-\-]/`, `/[\^\--]/`},
		{`/[^\-\-\-]/`, `/[^-\--]/`},
		{`/[^a-b\-]/`, `/[^a-b-]/`},
		{`/[^a-b\-\-]/`, `/[^a-b--]/`},
		{`/[^a-b\-\-\-]/`, `/[^a-b-\--]/`},
		{`/[^a\-\--\-\-\-]/`, `/[^a\-\-----]/`},

		// edge-cases
		{`let o=null;try{o=(o?.a).b||"FAIL"}catch(x){}console.log(o||"PASS")`, `let o=null;try{o=o?.a.b||"FAIL"}catch{}console.log(o||"PASS")`},
		{`1..a`, `1..a`},
		{`1.5.a`, `1.5.a`},
		{`1e4.a`, `1e4.a`},
		{`t0.a`, `t0.a`},
		{`for(;a < !--script;);`, `for(;a<! --script;);`},
		{`for(;a < /script>/;);`, `for(;a< /script>/;);`},
		{`a<<!--script`, `a<<! --script`},
		{`a<</script>/`, `a<< /script>/`},
		{`function f(a,b){a();for(const c of b){const b=0}}`, `function f(a,b){a();for(const c of b){const b=0}}`},
		{`return a,b,void 0`, `return a,b`},
		{`var arr=[];var slice=arr.slice;var concat=arr.concat;var push=arr.push;var indexOf=arr.indexOf;var class2type={};`, `var arr=[],slice=arr.slice,concat=arr.concat,push=arr.push,indexOf=arr.indexOf,class2type={}`},
		{`var arr=[];var class2type={};a=5;var rlocalProtocol=0`, `var arr=[],class2type={};a=5;var rlocalProtocol=0`},
		{`a=b;if(!o)return c;return d`, `return a=b,o?d:c`},

		// go-fuzz
		{`({"":a})`, `({"":a})`},
		{`a[""]`, `a[""]`},
		{`function f(){;}`, `function f(){}`},
		{`0xeb00000000`, `0xeb00000000`},
		{`export{a,}`, `export{a,}`},
		{`var D;var{U,W,W}=y`, `var{U,W,W}=y,D`},
		{`var A;var b=(function(){var e;})=c,d`, `var d,A,b=function(){var e}=c`},
		{`0xB_BBBbAbA`, `3149642426`},
		{`"\udFEb"`, `"\udFEb"`},

		// bugs
		{"var a=/\\s?auto?\\s?/i\nvar b;a,b", "var b,a=/\\s?auto?\\s?/i;a,b"},                           // #14
		{"false`string`", "(!1)`string`"},                                                               // #181
		{"x / /\\d+/.exec(s)[0]", "x/ /\\d+/.exec(s)[0]"},                                               // #183
		{`()=>{return{a}}`, `()=>({a})`},                                                                // #333
		{`()=>({a})`, `()=>({a})`},                                                                      // #333
		{`function f(){if(a){return 1}else if(b){return 2}return 3}`, `function f(){return a?1:b?2:3}`}, // #335
		{`new RegExp("\xAA\xB5")`, `new RegExp("\xAA\xB5")`},                                            // #341
		{`for(var a;;)a();var b=5`, `for(;;)a();var a,b=5`},                                             // #346
		{`if(e?0:n=1,o=2){o.a}`, `(e?0:n=1,o=2)&&o.a`},                                                  // #347
		{`const a=(a,b)=>({...a,b})`, `const a=(a,b)=>({...a,b})`},                                      // #369
		{`if(!a)debugger;`, `if(!a)debugger`},                                                           // #370
		{`export function a(b){b}`, `export function a(b){b}`},                                          // #375
		{`switch(a){case 0:b=c;d=e}`, `switch(a){case 0:b=c,d=e}`},                                      // #426
		{`if(a){const b=0}`, ``},                                                                        // #428
		{`()=>({a(){b=!b}})`, `()=>({a(){b=!b}})`},                                                      // #429
		{`var a=1;function f(){return 1}var{min,max}=Math;function g(){return 2}`, `a=1;function f(){return 1}var{min,max}=Math,a;function g(){return 2}`}, // #445
		{`const f=x=>void console.log(x)`, `const f=x=>void console.log(x)`},                                                                               // #463
		{`(function(){var a=b;var c=d.x,e=f.y})()`, `(function(){var a=b,c=d.x,e=f.y})()`},                                                                 // #472
		{`var a=1;g();a=2;let b=3`, `var a=1;g(),a=2;let b=3`},                                                                                             // #474
		{`if(!(0<1&&1<2)){throw new Error()}`, `if(!(0<1&&1<2))throw new Error`},                                                                           // #479
		{`class A{set x(e){}}`, `class A{set x(e){}}`},                                                                                                     // #481
		{`if(a){let b=c(d)}`, `a&&c(d)`},                                                                                                                   // #487
		{`var a=5;({});var b={c:()=>3}`, `var b,a=5;({},b={c:()=>3})`},                                                                                     // #494
		{`var a=5;({});var b=function(){3}`, `var b,a=5;({},b=function(){3})`},                                                                             // #494
		{`var a=5;({});var b=class{c(){3}}`, `var b,a=5;({},b=class{c(){3}})`},                                                                             // #494
		{`({});a={b(){3}}`, `({},a={b(){3}})`},                                                                                                             // #494
		{`export default function Foo(){a}Foo.prototype.bar=b`, `export default function Foo(){a}Foo.prototype.bar=b`},                                     // #525
		{`(e=1,e=2)`, `e=1,e=2`},                  // #528
		{`"\x00\x31 \0\u0000"`, `"\x001 \0\x00"`}, // #577
	}

	m := minify.New()
	o := Minifier{KeepVarNames: true, useAlphabetVarNames: true}
	for _, tt := range jsTests {
		t.Run(tt.js, func(t *testing.T) {
			r := bytes.NewBufferString(tt.js)
			w := &bytes.Buffer{}
			err := o.Minify(m, w, r, nil)
			test.Minify(t, tt.js, err, w.String(), tt.expected)
		})
	}
}

func TestJSVarRenaming(t *testing.T) {
	jsTests := []struct {
		js       string
		expected string
	}{
		{`x=function(){var name;name}`, `x=function(){var a;a}`},
		{`x=function(){var once,twice;once,twice++}`, `x=function(){var a,b;a,b++}`},
		{`x=function(){try{var x;x}catch(y){x,y}}`, `x=function(){try{var a;a}catch(b){a,b}}`},
		{`x=function(){try{var x;x}catch(x){x}}`, `x=function(){try{var a;a}catch(a){a}}`},
		{`x=function(){function name(){}}`, `x=function(){function a(){}}`},
		{`x=function name(){}`, `x=function(){}`},
		{`x=function(){let a;{let b;b,a}}`, `x=function(){let a;{let b;b,a}}`},
		//{`x=function(){let a;{let b;a}}`, `x=function(){let a;a}`}, // TODO: b unused
		{`x=function({foo, bar}){}`, `x=function({foo:a,bar:b}){}`},
		{`x=function(){class Wheel{}}`, `x=function(){class a{}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2}}`, `x=function(){function a(a,b){return a,b}}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return arg1}`, `x=function(){function a(a,b){return a,b}return arg1}`},
		{`x=function(){function name(arg1, arg2){return arg1, arg2} return a}`, `x=function(){function b(a,b){return a,b}return a}`},
		{`x=function(){function add(l,r){return add(l,r)}function nadd(l,r){return-add(l,r)}}`, `x=function(){function a(b,c){return a(b,c)}function b(b,c){return-a(b,c)}}`},
		{`function a(){var b;b}`, `function a(){var a;a}`},
		{`!function(){x=function(){return fun()};var fun=function(){return 0}}`, `!function(){x=function(){return a()};var a=function(){return 0}}`},
		{`!function(){var x=function(){return y};const y=5;x,y}`, `!function(){var b=function(){return a};const a=5;b,a}`},
		{`!function(){if(1){const x=5;x;5}var y=function(){return x};y}`, `!function(){if(1){const a=5;a,5}var a=function(){return x};a}`},
		{`!function(){var x=function(){return y};x;if(1){const y=5;y;5}}`, `!function(){var a=function(){return y};if(a,1){const a=5;a,5}}`},
		{`!function(){var x=function(){return y};x;if(z)var y=5}`, `!function(){var a,b=function(){return a};b,z&&(a=5)}`},
		{`!function(){var x=function(){return y};x;if(z){var y=5;5}}`, `!function(){var a,b=function(){return a};b,z&&(a=5,5)}`},
		{`!function(){var x,y,z=(x,y)=>x+y;x,y,z}`, `!function(){var a,b,c=(a,b)=>a+b;a,b,c}`},
		{`!function(){var await;print({await});}`, `!function(){var a;print({await:a})}`},
		{`function a(){var name; return {name}}`, `function a(){var a;return{name:a}}`},
		{`function a(){try{}catch(arg){arg}}`, `function a(){try{}catch(a){a}}`},
		{`function a(){var name,z;z;try{}catch(name){var name}}`, `function a(){var a,b;b;try{}catch{}}`},
		{`function a(){var name,z;z;try{}catch(arg){var name}}`, `function a(){var a,b;b;try{}catch{}}`},
		{`function a(b){function c(d){b[d]}}`, `function a(a){function b(b){a[b]}}`},
		{`function r(o){function l(t){if(!z[t]){if(!o[t]);}}}`, `function r(a){function b(b){z[b]||!a[b]}}`},
		{`!function(a){a;for(var b=0;;);};var c;var d;`, `!function(a){a;for(var b=0;;);};var c,d`},
		{`!function(){var b;b;{(T=x),T}{var T}}`, `!function(){var a,b;b,a=x,a}`},
		{`var T;T;!function(){var b;b;{(T=x),T}{var T}}`, `var T;T,!function(){var a,b;b,a=x,a}`},
		{`!function(){let a=b,b=c,c=d,d=e,e=f,f=g,g=h,h=a,j;for(let i=0;;)j=4}`, `!function(){let a=b,b=c,c=d,d=e,e=f,f=g,g=h,h=a,i;for(let a=0;;)i=4}`},
		{`function a(){var name;with(z){name}} function b(){var name;name}`, `function a(){var name;with(z)name}function b(){var a;a}`},
		{`!function(){var name;{name;!function(){name;var other;other}}}`, `!function(){var a;a,!function(){a;var b;b}}`},
		{`name=function(){var a001,a002,a003,a004,a005,a006,a007,a008,a009,a010,a011,a012,a013,a014,a015,a016,a017,a018,a019,a020,a021,a022,a023,a024,a025,a026,a027,a028,a029,a030,a031,a032,a033,a034,a035,a036,a037,a038,a039,a040,a041,a042,a043,a044,a045,a046,a047,a048,a049,a050,a051,a052,a053,a054,a055,a056,a057,a058,a059,a060,a061,a062,a063,a064,a065,a066,a067,a068,a069,a070,a071,a072,a073,a074,a075,a076,a077,a078,a079,a080,a081,a082,a083,a084,a085,a086,a087,a088,a089,a090,a091,a092,a093,a094,a095,a096,a097,a098,a099,a100,a101,a102,a103,a104,a105,a106,a107,a108,a109,a110,a111,a112,a113,a114,a115,a116,a117,a118,a119;a001,a002,a003,a004,a005,a006,a007,a008,a009,a010,a011,a012,a013,a014,a015,a016,a017,a018,a019,a020,a021,a022,a023,a024,a025,a026,a027,a028,a029,a030,a031,a032,a033,a034,a035,a036,a037,a038,a039,a040,a041,a042,a043,a044,a045,a046,a047,a048,a049,a050,a051,a052,a053,a054,a055,a056,a057,a058,a059,a060,a061,a062,a063,a064,a065,a066,a067,a068,a069,a070,a071,a072,a073,a074,a075,a076,a077,a078,a079,a080,a081,a082,a083,a084,a085,a086,a087,a088,a089,a090,a091,a092,a093,a094,a095,a096,a097,a098,a099,a100,a101,a102,a103,a104,a105,a106,a107,a108,a109,a110,a111,a112,a113,a114,a115,a116,a117,a118,a119}`,
			`name=function(){var a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,_,$,aa,ba,ca,da,ea,fa,ga,ha,ia,ja,ka,la,ma,na,oa,pa,qa,ra,sa,ta,ua,va,wa,xa,ya,za,Aa,Ba,Ca,Da,Ea,Fa,Ga,Ha,Ia,Ja,Ka,La,Ma,Na,Oa,Pa,Qa,Ra,Sa,Ta,Ua,Va,Wa,Xa,Ya,Za,_a,$a,ab,bb,cb,db,eb,fb,gb,hb,ib,jb,kb;a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,_,$,aa,ba,ca,da,ea,fa,ga,ha,ia,ja,ka,la,ma,na,oa,pa,qa,ra,sa,ta,ua,va,wa,xa,ya,za,Aa,Ba,Ca,Da,Ea,Fa,Ga,Ha,Ia,Ja,Ka,La,Ma,Na,Oa,Pa,Qa,Ra,Sa,Ta,Ua,Va,Wa,Xa,Ya,Za,_a,$a,ab,bb,cb,db,eb,fb,gb,hb,ib,jb,kb}`}, // 'as' is a keyword
		{`a=>{for(let b of c){b,a;{var d}}}`, `a=>{for(let d of c){d,a;var b}}`}, // #334
		//{`({x,y,z})=>x+y+z`, `({x,y,z})=>x+y+z`},
		{`function f(a){let b=0;if(a===0){return 0}else{let b=3;return b}}`, `function f(a){let c=0;if(a===0)return 0;let b=3;return b}`}, // #405
		{`!function(a){let b=0;if(a===0){return 0}else{let b=3;return b}}`, `!function(a){let c=0;if(a===0)return 0;let b=3;return b}`},   // #405
		{`a=>{let b=0;if(a===0){return 0}else{let b=3;return b}}`, `a=>{let c=0;if(a===0)return 0;let b=3;return b}`},                     // #405
		{`{let b=0;if(a===0){return 0}else{let b=3;return b}}`, `{let c=0;if(a===0)return 0;let b=3;return b}`},                           // #405
		{`class x{f(a){let b=0;if(a===0){return 0}else{let b=3;return b}}}`, `class x{f(a){let c=0;if(a===0)return 0;let b=3;return b}}`}, // #405
		{`for(;;){let b=0;if(a===0){return 0}else{let b=3;return b}}`, `for(;;){let c=0;if(a===0)return 0;let b=3;return b}`},             // #405
		{`try{let b=0;if(a===0){return 0}else{let b=3;return b}}catch{}`, `try{let c=0;if(a===0)return 0;let b=3;return b}catch{}`},       // #405
		{`let a=0;switch(a){case 0:let b=1;case 1:let c=2}`, `let a=0;switch(a){case 0:let a=1;case 1:let b=2}`},
		{`({a:b=1}={})=>b`, `({a=1}={})=>a`}, // #422
		{`()=>{var a;if(x){const b=0;while(true);}}`, `()=>{if(x){const b=0;for(var a;!0;);}}`},
		{`(e,s)=>{e=>0,s(e(s))}`, `(a,b)=>{a=>0,b(a(b))}`}, // #469
		{`()=>{var c;try {a} catch(b) {c}}`, `()=>{var b;try{a}catch{b}}`},

		// go-fuzz
		{`var ÆÆ,ÆÆ=t;var ÆÆ=v,a=ÿ`, `var ÆÆ=t,ÆÆ=v,a=ÿ`},
	}

	m := minify.New()
	o := Minifier{useAlphabetVarNames: true}
	for _, tt := range jsTests {
		t.Run(tt.js, func(t *testing.T) {
			r := bytes.NewBufferString(tt.js)
			w := &bytes.Buffer{}
			err := o.Minify(m, w, r, nil)
			test.Minify(t, tt.js, err, w.String(), tt.expected)
		})
	}
}

func TestJSVersion(t *testing.T) {
	versions := []int{2022, 2020, 2019, 2018}

	jsTests := []struct {
		version int
		js      string
		before  string
		after   string
	}{
		{2020, `a==null?b:a`, `a==null?b:a`, `a??b`},
		{2019, `try{}catch(a){}`, `try{}catch(a){}`, `try{}catch{}`},
	}

	m := minify.New()
	for _, tt := range jsTests {
		for _, version := range versions {
			t.Run(fmt.Sprintf("%d/%v", version, tt.js), func(t *testing.T) {
				r := bytes.NewBufferString(tt.js)
				w := &bytes.Buffer{}
				o := Minifier{KeepVarNames: true, useAlphabetVarNames: true, Version: version}
				err := o.Minify(m, w, r, nil)
				if version < tt.version {
					test.Minify(t, tt.js, err, w.String(), tt.before)
				} else {
					test.Minify(t, tt.js, err, w.String(), tt.after)
				}
			})
		}
	}
}

func TestReaderError(t *testing.T) {
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	m := minify.New()
	err := Minify(m, w, r, nil)
	test.T(t, err, test.ErrPlain)
}

func TestWriterError(t *testing.T) {
	r := bytes.NewBufferString("a")
	w := test.NewErrorWriter(0)
	m := minify.New()
	err := Minify(m, w, r, nil)
	test.T(t, err, test.ErrPlain)
}

func TestRenamerIndices(t *testing.T) {
	renamer := newRenamer(true, true)
	for _, i := range []int{0, 1, 2, 53, 54, 55, 117, 118} {
		name := renamer.getName([]byte{' '}, i)
		j := renamer.getIndex(name)
		test.T(t, j, i, string(name))
	}
	for i := 0; i < 100000; i++ {
		name := renamer.getName([]byte{' '}, i)
		j := renamer.getIndex(name)
		test.T(t, j, i, string(name))
	}
}

func BenchmarkJQuery(b *testing.B) {
	m := minify.New()
	buf, err := ioutil.ReadFile("../benchmarks/sample_jquery.js")
	if err != nil {
		panic(err)
	}
	for j := 0; j < 10; j++ {
		b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
			b.SetBytes(int64(len(buf)))
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				runtime.GC()
				r := buffer.NewReader(parse.Copy(buf))
				w := buffer.NewWriter(make([]byte, 0, len(buf)))
				b.StartTimer()

				if err := Minify(m, w, r, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("application/javascript", Minify)

	if err := m.Minify("application/javascript", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
