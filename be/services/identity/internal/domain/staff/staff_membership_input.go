package staff

type SetOrganizationStaffInput struct{ OrganizationID, UserID, Role, Status string }
type SetVenueStaffInput struct{ VenueID, UserID, Status string }

type AddOrganizationStaffInput struct{ OrganizationID, Email, Role string }
type RemoveOrganizationStaffInput struct{ OrganizationID, UserID string }
