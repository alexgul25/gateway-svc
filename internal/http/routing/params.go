package routing

const (
	ParamUserID     = "userID"
	ParamFolloweeID = "followeeID"

	PathUsers            = "/users"
	PathAuthLogin        = "/auth/login"
	PathUsersMe          = "/users/me"
	PathSubscriptions    = "/subscriptions"
	PathSubscriptionByID = "/subscriptions/{" + ParamFolloweeID + "}"
	PathUsersMeFollowers = "/users/me/followers"
	PathUserFollowers    = "/users/{" + ParamUserID + "}/followers"

	PathPlaces     = "/places"
	PathMyPlaces   = "/users/me/places"
	PathUserPlaces = "/users/{" + ParamUserID + "}/places"
)
