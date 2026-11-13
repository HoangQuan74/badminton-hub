package exception

var VenueErrors = ErrorConfig{NotFoundCode: "VENUE_NOT_FOUND", AlreadyExistsCode: "VENUE_SLUG_ALREADY_EXISTS"}
var CourtErrors = ErrorConfig{NotFoundCode: "COURT_NOT_FOUND", AlreadyExistsCode: "COURT_NAME_ALREADY_EXISTS"}
var ImageErrors = ErrorConfig{NotFoundCode: "VENUE_IMAGE_NOT_FOUND", UnavailableCode: "DEPENDENCY_UNAVAILABLE"}
