package java

import (
	"slices"
	"testing"
)

func TestParseJavaClassNames(t *testing.T) {
	// Parse a class where the class/enum keywords are used in comments
	classNames, err := parseJavaClassNames(
		"com/abc/def/xyz/HelloWorld.java",
		`
		package com.abc.def.xyz;

		public class HelloWorld {
			/**
			 * This is a class that prints "Hello, World!" to the console.
			 */
			public static void main(String[] args) {
				System.out.println("Hello, World!");
			}

			// This is an enum which is used to represent colors.
			public static enum Color {
				RED,
				GREEN,
				BLUE
			}

			public static abstract class SomeAbstractClass {
				// This is a method that does something.
				public abstract void doSomething();
			}
		}
		`)
	if err != nil {
		t.Errorf("ParseJavaClassNames failed: %v", err)
	}
	if len(classNames) != 3 {
		t.Errorf("ParseJavaClassNames returned unexpected number of class names: %d", len(classNames))
	}
	if !slices.Contains(classNames, "com.abc.def.xyz.HelloWorld") {
		t.Errorf("missing top-level class name: com.abc.def.xyz.HelloWorld")
	}
	if !slices.Contains(classNames, "com.abc.def.xyz.HelloWorld.Color") {
		t.Errorf("missing inner class name: com.abc.def.xyz.HelloWorld.Color")
	}
	if !slices.Contains(classNames, "com.abc.def.xyz.HelloWorld.SomeAbstractClass") {
		t.Errorf("missing abstract class name: com.abc.def.xyz.HelloWorld.SomeAbstractClass")
	}
}
