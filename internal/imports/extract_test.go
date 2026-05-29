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

func TestExtractIgnoresCommentsAndStringLiterals(t *testing.T) {
	source := []byte(`
// import { Commented } from "@/modules/commented";
/*
export { BlockCommented } from "@/modules/block-commented";
*/
const sample = "import { InString } from '@/modules/string-literal'";
const template = ` + "`require('@/modules/template-literal')`" + `;
import { Real } from "@/modules/real";
`)

	statements := Extract(source)
	if len(statements) != 1 {
		t.Fatalf("expected only one real import, got %#v", statements)
	}
	if statements[0].Path != "@/modules/real" {
		t.Fatalf("expected real import, got %#v", statements[0])
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
