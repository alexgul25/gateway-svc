package routing

const (
	ParamUserID     = "userID"
	ParamFolloweeID = "followeeID"

	PathUsers            = "/users"
	PathAuthLogin        = "/auth/login"
	PathUsersMe          = "/users/me"
	PathSubscriptions    = "/subscriptions"
	PathSubscriptionByID = "/subscriptions/{followeeID}"
	PathUsersMeFollowers = "/users/me/followers"
	PathUserFollowers    = "/users/{userID}/followers"
)
