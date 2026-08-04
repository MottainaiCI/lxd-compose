/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package helpers

type ChannelError struct {
	Error   error
	Closure interface{}
}
