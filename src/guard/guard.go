package guard

import "github.com/nexgou/server/src/common"

// Execute runs a list of guards against the current context in order.
// Returns a ForbiddenException if any guard denies access, or the guard's own error.
func Execute(ctx *common.Context, guards ...common.Guard) error {
	for _, g := range guards {
		ok, err := g.CanActivate(ctx)
		if err != nil {
			return err
		}
		if !ok {
			return common.NewForbiddenException("Forbidden")
		}
	}
	return nil
}
