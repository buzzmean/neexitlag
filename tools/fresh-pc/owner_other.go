//go:build !linux

package main

func rootPrivateCA(string) bool { return false }
