package ride

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wavesplatform/gowaves/pkg/ride/serialization"
)

func c(values ...rideType) []rideType {
	return values
}

func TestSimpleScriptsCompilation(t *testing.T) {
	for _, test := range []struct {
		comment   string
		source    string
		code      string
		constants []rideType
	}{
		{`V1: true`, "AQa3b8tH", "0400", nil},
		{`V3: let x = 1; true`, "AwQAAAABeAAAAAAAAAAAAQbtAkXn", "040002000001", c(rideInt(1))},
		{`V3: let x = "abc"; true`, "AwQAAAABeAIAAAADYWJjBrpUkE4=", "040002000001", c(rideString("abc"))},
		{`V3: func A() = 1; func B() = 2; true`, "AwoBAAAAAUEAAAAAAAAAAAAAAAABCgEAAAABQgAAAAAAAAAAAAAAAAIG+N0aQQ==",
			"04000200010102000001", c(rideInt(1), rideInt(2))},
		{`V3: func A() = 1; func B() = 2; A() != B()`, "AwoBAAAAAUEAAAAAAAAAAAAAAAABCgEAAAABQgAAAAAAAAAAAAAAAAIJAQAAAAIhPQAAAAIJAQAAAAFBAAAAAAkBAAAAAUIAAAAAv/Pmkg==",
			"0a001400000a001000000900010002000200010102000001", c(rideInt(1), rideInt(2))},
		{`V1: let i = 1; let s = "string"; toString(i) == s`, "AQQAAAABaQAAAAAAAAAAAQQAAAABcwIAAAAGc3RyaW5nCQAAAAAAAAIJAAGkAAAAAQUAAAABaQUAAAABcwIsH74=",
			"0c001509002700010c00110900030002000200010102000001", c(rideInt(1), rideString("string"))},
		{`V3: if true then if true then true else false else false`, "AwMGAwYGBwdYjCji",
			"04070013030407000e03040600100305060015030500", nil},
		{`V3: if (true) then {let r = true; r} else {let r = false; r}`, "AwMGBAAAAAFyBgUAAAABcgQAAAABcgcFAAAAAXJ/ok0E",
			"0407000b030c001006000f030c00120004010501", nil},
		{`V3: if (let a = 1; a == 0) then {let a = 2; a == 0} else {let a = 0; a == 0}`, "AwMEAAAAAWEAAAAAAAAAAAEJAAAAAAAAAgUAAAABYQAAAAAAAAAAAAQAAAABYQAAAAAAAAAAAgkAAAAAAAACBQAAAAFhAAAAAAAAAAAABAAAAAFhAAAAAAAAAAAACQAAAAAAAAIFAAAAAWEAAAAAAAAAAAB3u9Yb",
			"0c002a020001090003000207001d030c002e0200030900030002060029030c0032020005090003000200020000010200020102000401", c(rideInt(1), rideInt(0), rideInt(2), rideInt(0), rideInt(0), rideInt(0))},
		{`let a = 1; let b = a; let c = b; a == c`,
			"AwQAAAABYQAAAAAAAAAAAQQAAAABYgUAAAABYQQAAAABYwUAAAABYgkAAAAAAAACBQAAAAFhBQAAAAFjUFI1Og==",
			"0c00140c000c0900030002000c0010010c00140102000001", c(rideInt(1))},
		{`let x = addressFromString("3PJaDyprvekvPXPuAtxrapacuDJopgJRaU3"); let a = x; let b = a; let c = b; let d = c; let e = d; let f = e; f == e`,
			"AQQAAAABeAkBAAAAEWFkZHJlc3NGcm9tU3RyaW5nAAAAAQIAAAAjM1BKYUR5cHJ2ZWt2UFhQdUF0eHJhcGFjdURKb3BnSlJhVTMEAAAAAWEFAAAAAXgEAAAAAWIFAAAAAWEEAAAAAWMFAAAAAWIEAAAAAWQFAAAAAWMEAAAAAWUFAAAAAWQEAAAAAWYFAAAAAWUJAAAAAAAAAgUAAAABZgUAAAABZS5FHzs=",
			"0c000c0c00100900030002000c0010010c0014010c0018010c001c010c0020010c00240102000009003a000101", c(rideString("3PJaDyprvekvPXPuAtxrapacuDJopgJRaU3"))},
		{`V3: let x = { let y = 1; y == 0 }; let y = { let z = 2; z == 0 } x == y`,
			"AwQAAAABeAQAAAABeQAAAAAAAAAAAQkAAAAAAAACBQAAAAF5AAAAAAAAAAAABAAAAAF5BAAAAAF6AAAAAAAAAAACCQAAAAAAAAIFAAAAAXoAAAAAAAAAAAAJAAAAAAAAAgUAAAABeAUAAAABedn8HVg=",
			"0c00200c001409000300020002000001020002010c00100200030900030002010c000c020001090003000201", c(rideInt(1), rideInt(0), rideInt(2), rideInt(0))},
		{`V3: let z = 0; let a = {let b = 1; b == z}; let b = {let c = 2; c == z}; a == b`,
			"AwQAAAABegAAAAAAAAAAAAQAAAABYQQAAAABYgAAAAAAAAAAAQkAAAAAAAACBQAAAAFiBQAAAAF6BAAAAAFiBAAAAAFjAAAAAAAAAAACCQAAAAAAAAIFAAAAAWMFAAAAAXoJAAAAAAAAAgUAAAABYQUAAAABYnau3I8=",
			"0c00200c001409000300020002000101020002010c00100c002c0900030002010c000c0c002c09000300020102000001", c(rideInt(0), rideInt(1), rideInt(2))},
		{`V3: func abs(i:Int) = if (i >= 0) then i else -i; abs(-10) == 10`, "AwoBAAAAA2FicwAAAAEAAAABaQMJAABnAAAAAgUAAAABaQAAAAAAAAAAAAUAAAABaQkBAAAAAS0AAAABBQAAAAFpCQAAAAAAAAIJAQAAAANhYnMAAAABAP/////////2AAAAAAAAAAAKmp8BWw==",
			"0200010a001100010200020900030002000d000002000009000d0002070026030d000006002f030d0000090002000101", c(rideInt(0), rideInt(-10), rideInt(10))},
		{`V3: if (true) then {if (false) then {func XX() = true; XX()} else {func XX() = false; XX()}} else {if (true) then {let x = false; x} else {let x = true; x}}`,
			"AwMGAwcKAQAAAAJYWAAAAAAGCQEAAAACWFgAAAAACgEAAAACWFgAAAAABwkBAAAAAlhYAAAAAAMGBAAAAAF4BwUAAAABeAQAAAABeAYFAAAAAXgYYeMi",
			"0407001b0305070012030a002c0000060018030a002e000006002b0304070027030c003006002b030c0032000401050105010401", nil},
		{`tx.sender == Address(base58'11111111111111111')`, "AwkAAAAAAAACCAUAAAACdHgAAAAGc2VuZGVyCQEAAAAHQWRkcmVzcwAAAAEBAAAAEQAAAAAAAAAAAAAAAAAAAAAAWc7d/w==",
			"0b00180800000200010900510001090003000200", c(rideString("sender"), rideBytes{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})},
		{`func b(x: Int) = {func a(y: Int) = x + y; a(1) + a(2)}; b(2) + b(3) == 0`, "AwoBAAAAAWIAAAABAAAAAXgKAQAAAAFhAAAAAQAAAAF5CQAAZAAAAAIFAAAAAXgFAAAAAXkJAABkAAAAAgkBAAAAAWEAAAABAAAAAAAAAAABCQEAAAABYQAAAAEAAAAAAAAAAAIJAAAAAAAAAgkAAGQAAAACCQEAAAABYgAAAAEAAAAAAAAAAAIJAQAAAAFiAAAAAQAAAAAAAAAAAwAAAAAAAAAAAPsZlhQ=",
			"0200020a002a00010200030a002a000109000500020200040900030002000d00000d00000900050002010200000a001e00010200010a001e0001090005000201", c(rideInt(1), rideInt(2), rideInt(2), rideInt(3), rideInt(0))},
		{`func first(a: Int, b: Int) = {let x = a + b; x}; first(1, 2) == 0`, "AwoBAAAABWZpcnN0AAAAAgAAAAFhAAAAAWIEAAAAAXgJAABkAAAAAgUAAAABYQUAAAABYgUAAAABeAkAAAAAAAACCQEAAAAFZmlyc3QAAAACAAAAAAAAAAABAAAAAAAAAAACAAAAAAAAAAAAm+QHtw==",
			"0200000200010a002000020200020900030002000d00000d00010900050002010c001401", c(rideInt(1), rideInt(2), rideInt(0))},
		{`func A(x: Int, y: Int) = {let r = x + y; r}; func B(x: Int, y: Int) = {let r = A(x, y); r}; B(1, 2) == 3`, "AwoBAAAAAUEAAAACAAAAAXgAAAABeQQAAAABcgkAAGQAAAACBQAAAAF4BQAAAAF5BQAAAAFyCgEAAAABQgAAAAIAAAABeAAAAAF5BAAAAAFyCQEAAAABQQAAAAIFAAAAAXgFAAAAAXkFAAAAAXIJAAAAAAAAAgkBAAAAAUIAAAACAAAAAAAAAAABAAAAAAAAAAACAAAAAAAAAAADSAdb8g==",
			"0200000200010a002c00020200020900030002000d00000d00010900050002010d00000d00010a00300002010c0020010c001401", c(rideInt(1), rideInt(2), rideInt(3))},
		{`func f1(a: Int, b: Int) = a + b; func f2(a: Int, b: Int) = a - b; f2(f1(1, 2), 3) == 0`, "AwoBAAAAAmYxAAAAAgAAAAFhAAAAAWIJAABkAAAAAgUAAAABYQUAAAABYgoBAAAAAmYyAAAAAgAAAAFhAAAAAWIJAABlAAAAAgUAAAABYQUAAAABYgkAAAAAAAACCQEAAAACZjIAAAACCQEAAAACZjEAAAACAAAAAAAAAAABAAAAAAAAAAACAAAAAAAAAAADAAAAAAAAAAAALZ/RdA==",
			"0200000200010a002800020200020a001c00020200030900030002000d00000d000109000b0002010d00000d0001090005000201", c(rideInt(1), rideInt(2), rideInt(3), rideInt(0))},
		{`func f1(a: Int, b: Int) = a + b; func f2(a: Int, b: Int) = a - b; let x = f1(1, 2); f2(x, 3) == 0`, "AwoBAAAAAmYxAAAAAgAAAAFhAAAAAWIJAABkAAAAAgUAAAABYQUAAAABYgoBAAAAAmYyAAAAAgAAAAFhAAAAAWIJAABlAAAAAgUAAAABYQUAAAABYgQAAAABeAkBAAAAAmYxAAAAAgAAAAAAAAAAAQAAAAAAAAAAAgkAAAAAAAACCQEAAAACZjIAAAACBQAAAAF4AAAAAAAAAAADAAAAAAAAAAAAr1ooAg==",
			"0c00140200020a002000020200030900030002000200000200010a002c0002010d00000d000109000b0002010d00000d0001090005000201", c(rideInt(1), rideInt(2), rideInt(3), rideInt(0))},
		{`func f1(a: Int, b: Int) = a + b; func f2(a: Int, b: Int) = b; f2(f1(1, 2), 3) == 3`, "AwoBAAAAAmYxAAAAAgAAAAFhAAAAAWIJAABkAAAAAgUAAAABYQUAAAABYgoBAAAAAmYyAAAAAgAAAAFhAAAAAWIFAAAAAWIJAAAAAAAAAgkBAAAAAmYyAAAAAgkBAAAAAmYxAAAAAgAAAAAAAAAAAQAAAAAAAAAAAgAAAAAAAAAAAwAAAAAAAAAAA1cKYN4=",
			"0200000200010a002000020200020a001c00020200030900030002000d0001010d00000d0001090005000201", c(rideInt(1), rideInt(2), rideInt(3), rideInt(3))},
		{`func f1(a: Int, b: Int) = a + b; func f2(a: Int, b: Int) = b; let x = f1(1, 2); f2(x, 3) == 3`, "AwoBAAAAAmYxAAAAAgAAAAFhAAAAAWIJAABkAAAAAgUAAAABYQUAAAABYgoBAAAAAmYyAAAAAgAAAAFhAAAAAWIFAAAAAWIEAAAAAXgJAQAAAAJmMQAAAAIAAAAAAAAAAAEAAAAAAAAAAAIJAAAAAAAAAgkBAAAAAmYyAAAAAgUAAAABeAAAAAAAAAAAAwAAAAAAAAAAA6avbPE=",
			"0c00140200020a002000020200030900030002000200000200010a00240002010d0001010d00000d0001090005000201", c(rideInt(1), rideInt(2), rideInt(3), rideInt(3))},
		{`let x = 1; func add(i: Int) = i + 1; add(x) == 2`, "AwQAAAABeAAAAAAAAAAAAQoBAAAAA2FkZAAAAAEAAAABaQkAAGQAAAACBQAAAAFpAAAAAAAAAAABCQAAAAAAAAIJAQAAAANhZGQAAAABBQAAAAF4AAAAAAAAAAACfr6U6w==",
			"0c001d0a001100010200020900030002000d000002000109000500020102000001", c(rideInt(1), rideInt(1), rideInt(2))},
		{`let b = base16'0000000000000001'; func add(b: ByteVector) = toInt(b) + 1; add(b) == 2`, "AwQAAAABYgEAAAAIAAAAAAAAAAEKAQAAAANhZGQAAAABAAAAAWIJAABkAAAAAgkABLEAAAABBQAAAAFiAAAAAAAAAAABCQAAAAAAAAIJAQAAAANhZGQAAAABBQAAAAFiAAAAAAAAAAACX00biA==",
			"0c00220a001100010200020900030002000d0000090020000102000109000500020102000001", c(rideBytes{0, 0, 0, 0, 0, 0, 0, 1}, rideInt(1), rideInt(2))},
		{`let b = base16'0000000000000001'; func add(v: ByteVector) = toInt(v) + 1; add(b) == 2`, "AwQAAAABYgEAAAAIAAAAAAAAAAEKAQAAAANhZGQAAAABAAAAAXYJAABkAAAAAgkABLEAAAABBQAAAAF2AAAAAAAAAAABCQAAAAAAAAIJAQAAAANhZGQAAAABBQAAAAFiAAAAAAAAAAACI7gYxg==",
			"0c00220a001100010200020900030002000d0000090020000102000109000500020102000001", c(rideBytes{0, 0, 0, 0, 0, 0, 0, 1}, rideInt(1), rideInt(2))},
		{`let b = base16'0000000000000001'; func add(v: ByteVector) = toInt(b) + 1; add(b) == 2`, "AwQAAAABYgEAAAAIAAAAAAAAAAEKAQAAAANhZGQAAAABAAAAAXYJAABkAAAAAgkABLEAAAABBQAAAAFiAAAAAAAAAAABCQAAAAAAAAIJAQAAAANhZGQAAAABBQAAAAFiAAAAAAAAAAAChRvwnQ==",
			"0c00220a001100010200020900030002000c0022090020000102000109000500020102000001", c(rideBytes{0, 0, 0, 0, 0, 0, 0, 1}, rideInt(1), rideInt(2))},
	} {
		src, err := base64.StdEncoding.DecodeString(test.source)
		require.NoError(t, err, test.comment)

		tree, err := serialization.Parse(src)
		require.NoError(t, err, test.comment)
		assert.NotNil(t, tree, test.comment)

		rideScript, err := Compile(tree)
		require.NoError(t, err, test.comment)
		assert.NotNil(t, rideScript, test.comment)
		script, ok := rideScript.(*SimpleScript)
		require.True(t, ok, test.comment)

		code := hex.EncodeToString(script.Code)
		assert.Equal(t, test.code, code, test.comment)
		assert.ElementsMatch(t, test.constants, script.Constants, test.comment)
	}
}

func TestDAppScriptsCompilation(t *testing.T) {
	for _, test := range []struct {
		comment   string
		source    string
		code      string
		constants []rideType
		entries   map[string]callable
	}{
		{`@Verifier(tx) func verify() = false`, "AAIDAAAAAAAAAAIIAQAAAAAAAAAAAAAAAQAAAAJ0eAEAAAAGdmVyaWZ5AAAAAAcysh6J",
			"0500", nil, map[string]callable{"": {0, "tx"}}},
		{`let a = 1\n@Verifier(tx) func verify() = false`, "AAIDAAAAAAAAAAIIAQAAAAEAAAAAAWEAAAAAAAAAAAEAAAAAAAAAAQAAAAJ0eAEAAAAGdmVyaWZ5AAAAAAdVrdkQ",
			"020000010500", c(rideInt(1)), map[string]callable{"": {4, "tx"}}},
		{`let a = 1\nfunc inc(v: Int) = {v + 1}\n@Verifier(tx) func verify() = false`, "AAIDAAAAAAAAAAIIAQAAAAIAAAAAAWEAAAAAAAAAAAEBAAAAA2luYwAAAAEAAAABdgkAAGQAAAACBQAAAAF2AAAAAAAAAAABAAAAAAAAAAEAAAACdHgBAAAABnZlcmlmeQAAAAAHDMc8rg==",
			"020000010d00000200010900050002010500", c(rideInt(1), rideInt(1)), map[string]callable{"": {16, "tx"}}},
		{`let a = 1\nfunc inc(v: Int) = {v + 1}\n@Verifier(tx) func verify() = inc(a) == 2`, "AAIDAAAAAAAAAAIIAQAAAAIAAAAAAWEAAAAAAAAAAAEBAAAAA2luYwAAAAEAAAABdgkAAGQAAAACBQAAAAF2AAAAAAAAAAABAAAAAAAAAAEAAAACdHgBAAAABnZlcmlmeQAAAAAJAAAAAAAAAgkBAAAAA2luYwAAAAEFAAAAAWEAAAAAAAAAAAJtD5WX",
			"020000010d00000200010900050002010c00000a00040001020002090003000200", c(rideInt(1), rideInt(1), rideInt(2)),
			map[string]callable{"": {16, "tx"}}},
		{`let a = 1\nlet b = 1\nfunc inc(v: Int) = {v + 1}\nfunc add(x: Int, y: Int) = {x + y}\n@Verifier(tx) func verify() = inc(a) == add(a, b)`, "AAIDAAAAAAAAAAIIAQAAAAQAAAAAAWEAAAAAAAAAAAEAAAAAAWIAAAAAAAAAAAEBAAAAA2luYwAAAAEAAAABdgkAAGQAAAACBQAAAAF2AAAAAAAAAAABAQAAAANhZGQAAAACAAAAAXgAAAABeQkAAGQAAAACBQAAAAF4BQAAAAF5AAAAAAAAAAEAAAACdHgBAAAABnZlcmlmeQAAAAAJAAAAAAAAAgkBAAAAA2luYwAAAAEFAAAAAWEJAQAAAANhZGQAAAACBQAAAAFhBQAAAAFiDbIkmw==",
			"02000001020001010d00000200020900050002010d00000d00010900050002010c00000a000800010c00000c00040a00140002090003000200",
			c(rideInt(1), rideInt(1), rideInt(1)),
			map[string]callable{"": {32, "tx"}}},
		{`let a = 1\nlet b = 1\nlet messages = ["INFO", "WARN"]\nfunc inc(v: Int) = {v + 1}\nfunc add(x: Int, y: Int) = {x + y}\nfunc msg(i: Int) = {messages[i]}\n@Verifier(tx) func verify() = if inc(a) == add(a, b) then throw(msg(a)) else throw(msg(b))`, "AAIDAAAAAAAAAAIIAQAAAAYAAAAAAWEAAAAAAAAAAAEAAAAAAWIAAAAAAAAAAAEAAAAACG1lc3NhZ2VzCQAETAAAAAICAAAABElORk8JAARMAAAAAgIAAAAEV0FSTgUAAAADbmlsAQAAAANpbmMAAAABAAAAAXYJAABkAAAAAgUAAAABdgAAAAAAAAAAAQEAAAADYWRkAAAAAgAAAAF4AAAAAXkJAABkAAAAAgUAAAABeAUAAAABeQEAAAADbXNnAAAAAQAAAAFpCQABkQAAAAIFAAAACG1lc3NhZ2VzBQAAAAFpAAAAAAAAAAEAAAACdHgBAAAABnZlcmlmeQAAAAADCQAAAAAAAAIJAQAAAANpbmMAAAABBQAAAAFhCQEAAAADYWRkAAAAAgUAAAABYQUAAAABYgkAAAIAAAABCQEAAAADbXNnAAAAAQUAAAABYQkAAAIAAAABCQEAAAADbXNnAAAAAQUAAAABYvi7IpM=",
			"02000001020001010200020200030b001609001e000209001e0002010d00000200040900050002010d00000d00010900050002010c00080d00000900320002010c00000a001c00010c00000c00040a00280002090003000207006c030c00000a00340001090028000106007a030c00040a00340001090028000100",
			c(rideInt(1), rideInt(1), rideString("INFO"), rideString("WARN"), rideInt(1)),
			map[string]callable{"": {64, "tx"}}},
		{`@Callable(i)func f() = {WriteSet([DataEntry("YYY", "XXX")]}`, "AAIDAAAAAAAAAAQIARIAAAAAAAAAAAEAAAABaQEAAAABZgAAAAAJAQAAAAhXcml0ZVNldAAAAAEJAARMAAAAAgkBAAAACURhdGFFbnRyeQAAAAICAAAAA1lZWQIAAAADWFhYBQAAAANuaWwAAAAAeFguLA==",
			"02000002000109005700020b001609001e0002090071000100", c(rideString("YYY"), rideString("XXX")),
			map[string]callable{"f": {0, "i"}}},
		{`@Callable(i)func f() = {let callerAddress = toBase58String(i.caller.bytes); WriteSet([DataEntry(callerAddress, "XXX")]}`, "AAIDAAAAAAAAAAQIARIAAAAAAAAAAAEAAAABaQEAAAABZgAAAAAEAAAADWNhbGxlckFkZHJlc3MJAAJYAAAAAQgIBQAAAAFpAAAABmNhbGxlcgAAAAVieXRlcwkBAAAACFdyaXRlU2V0AAAAAQkABEwAAAACCQEAAAAJRGF0YUVudHJ5AAAAAgUAAAANY2FsbGVyQWRkcmVzcwIAAAADWFhYBQAAAANuaWwAAAAAe3xtyw==",
			"0d000008000008000109003d0001010c000002000209005700020b001609001e0002090071000100",
			c(rideString("caller"), rideString("bytes"), rideString("XXX")),
			map[string]callable{"f": {15, "i"}}},
		{`let messages = ["INFO", "WARN"]\nfunc msg(i: Int) = {messages[i]}\n@Callable(i)func tellme(x: Int) = {WriteSet([DataEntry("m", msg(x))]}`, "AAIDAAAAAAAAAAcIARIDCgEBAAAAAgAAAAAIbWVzc2FnZXMJAARMAAAAAgIAAAAESU5GTwkABEwAAAACAgAAAARXQVJOBQAAAANuaWwBAAAAA21zZwAAAAEAAAABaQkAAZEAAAACBQAAAAhtZXNzYWdlcwUAAAABaQAAAAEAAAABaQEAAAAGdGVsbG1lAAAAAQAAAAF4CQEAAAAIV3JpdGVTZXQAAAABCQAETAAAAAIJAQAAAAlEYXRhRW50cnkAAAACAgAAAAFtCQEAAAADbXNnAAAAAQUAAAABeAUAAAADbmlsAAAAAO4TltI=",
			"0200000200010b001609001e000209001e0002010c00000d00000900320002010200020d00010a0014000109005700020b001609001e0002090071000100", c(rideString("INFO"), rideString("WARN"), rideString("m")),
			map[string]callable{"tellme": {32, "i"}}},
		{`let messages = ["INFO", "WARN"]\nfunc msg(i: Int) = {messages[i]}\n@Callable(i)func tellme(x: Int, y: Int) = {WriteSet([DataEntry("m", msg(x))]}`, "AAIDAAAAAAAAAAgIARIECgIBAQAAAAIAAAAACG1lc3NhZ2VzCQAETAAAAAICAAAABElORk8JAARMAAAAAgIAAAAEV0FSTgUAAAADbmlsAQAAAANtc2cAAAABAAAAAWkJAAGRAAAAAgUAAAAIbWVzc2FnZXMFAAAAAWkAAAABAAAAAWkBAAAABnRlbGxtZQAAAAIAAAABeAAAAAF5CQEAAAAIV3JpdGVTZXQAAAABCQAETAAAAAIJAQAAAAlEYXRhRW50cnkAAAACAgAAAAFtCQEAAAADbXNnAAAAAQUAAAABeAUAAAADbmlsAAAAAD8Tlfs=",
			"0200000200010b001609001e000209001e0002010c00000d00000900320002010200020d00010a0014000109005700020b001609001e0002090071000100", c(rideString("INFO"), rideString("WARN"), rideString("m")),
			map[string]callable{"tellme": {32, "i"}}},
		{`let a = 1; let messages = ["INFO", "WARN"]; func msg(i: Int) = {messages[i]}; @Callable(i)func tellme(x: Int) = {let m = msg(x); let callerAddress = toBase58String(i.caller.bytes); WriteSet([DataEntry(callerAddress + "-m", m)]}`, "AAIDAAAAAAAAAAcIARIDCgEBAAAAAwAAAAABYQAAAAAAAAAAAQAAAAAIbWVzc2FnZXMJAARMAAAAAgIAAAAESU5GTwkABEwAAAACAgAAAARXQVJOBQAAAANuaWwBAAAAA21zZwAAAAEAAAABaQkAAZEAAAACBQAAAAhtZXNzYWdlcwUAAAABaQAAAAEAAAABaQEAAAAGdGVsbG1lAAAAAQAAAAF4BAAAAAFtCQEAAAADbXNnAAAAAQUAAAABeAQAAAANY2FsbGVyQWRkcmVzcwkAAlgAAAABCAgFAAAAAWkAAAAGY2FsbGVyAAAABWJ5dGVzCQEAAAAIV3JpdGVTZXQAAAABCQAETAAAAAIJAQAAAAlEYXRhRW50cnkAAAACCQABLAAAAAIFAAAADWNhbGxlckFkZHJlc3MCAAAAAi1tBQAAAAFtBQAAAANuaWwAAAAAgveN3A==",
			"020000010200010200020b001609001e000209001e0002010c00040d00000900320002010d000008000308000409003d0001010d00010a00180001010c002402000509002d00020c003309005700020b001609001e0002090071000100", c(rideInt(1), rideString("INFO"), rideString("WARN"), rideString("caller"), rideString("bytes"), rideString("-m")),
			map[string]callable{"tellme": {60, "i"}}},
	} {
		src, err := base64.StdEncoding.DecodeString(test.source)
		require.NoError(t, err, test.comment)

		tree, err := serialization.Parse(src)
		require.NoError(t, err, test.comment)
		assert.NotNil(t, tree, test.comment)

		rideScript, err := Compile(tree)
		require.NoError(t, err, test.comment)
		assert.NotNil(t, rideScript, test.comment)
		script, ok := rideScript.(*DAppScript)
		require.True(t, ok, test.comment)

		code := hex.EncodeToString(script.Code)
		assert.Equal(t, test.code, code, test.comment)
		assert.ElementsMatch(t, test.constants, script.Constants, test.comment)
		assert.Equal(t, test.entries, script.EntryPoints, test.comment)
	}
}
