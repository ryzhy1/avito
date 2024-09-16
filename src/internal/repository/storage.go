package repository

import "fmt"

var (
	ErrNoAccessRights                      = fmt.Errorf("user is not responsible for the organization")
	ErrUserIsNotCreatorOrTenderWasNotFound = fmt.Errorf("tender was not found or user is not the creator")
	ErrTenderNotFound                      = fmt.Errorf("tender not found")
	ErrOrganizationNotFound                = fmt.Errorf("organization not found")
	ErrNoResponsible                       = fmt.Errorf("user is not responsible for the organization ")
	ErrNoAssociationWithOrganization       = fmt.Errorf("user is not associated with the organization")
	ErrUsernameFieldEmpty                  = fmt.Errorf("username field is empty")
	ErrBidNotFound                         = fmt.Errorf("bid not found")
	ErrNoPermission                        = fmt.Errorf("no permission")
	ErrTenderCloseFailed                   = fmt.Errorf("tender close failed")
	ErrVersionNotFound                     = fmt.Errorf("version not found")
	ErrReviewsNotFound                     = fmt.Errorf("reviews not found")
)
