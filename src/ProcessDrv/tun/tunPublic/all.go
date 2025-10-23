//go:build !darwin && !linux
// +build !darwin,!linux

package tunPublic

func GetGatewayByDefault() (string, string) {
	panic("implement me:!darwin,!linux")
}
