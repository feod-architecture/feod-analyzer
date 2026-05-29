package imports

import "testing"

func TestExtractSupportsImportShapes(t *testing.T) {
	source := []byte(`
import type { User } from "@/modules/user";
import {
  CheckoutFlow,
} from "@/modules/checkout";
import "@/global/styles.css";
const lazy = import("@/modules/lazy");
const common = require("@/common/date");
export { PaymentStep } from "@/modules/checkout";
export type { PaymentProps } from "@/modules/checkout";
`)

	statements := Extract(source)
	paths := map[string]bool{}
	for _, stmt := range statements {
		paths[stmt.Path] = true
	}

	for _, path := range []string{
		"@/modules/user",
		"@/modules/checkout",
		"@/global/styles.css",
		"@/modules/lazy",
		"@/common/date",
	} {
		if !paths[path] {
			t.Fatalf("expected import path %s in %#v", path, statements)
		}
	}
}

func TestExtractStarExports(t *testing.T) {
	source := []byte(`
export * from "./ui/UserCard";
export type * from "./model/types";
`)
	exports := ExtractStarExports(source)
	if len(exports) != 2 {
		t.Fatalf("expected 2 star exports, got %d", len(exports))
	}
	if exports[1].Path != "./model/types" || !exports[1].TypeOnly {
		t.Fatalf("unexpected type star export: %#v", exports[1])
	}
}
