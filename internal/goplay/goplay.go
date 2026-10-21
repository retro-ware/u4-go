// Package goplay is the game's side of goplay's folder layout: the game runs
// in games/u4/, the shared system/ folder is two levels up. The game has its
// own volume controls; goplay sets none.
package goplay

import "path/filepath"

const root = "../.."

// System is the path of a file in goplay's shared system folder.
func System(name string) string { return filepath.Join(root, "system", filepath.FromSlash(name)) }
