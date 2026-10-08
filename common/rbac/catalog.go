package rbac

func AuthRoutes() []Route {
	return []Route{
		Self(AudienceApp, "GET", "/api/"),
		Self(AudienceApp, "POST", "/api/logout"),
		Self(AudienceConsole, "GET", "/api/"),
		Self(AudienceConsole, "POST", "/api/logout"),
		Self(AudienceApp, "GET", "/api/rbac/menus"),
		Self(AudienceApp, "GET", "/api/rbac/roles"),
		Self(AudienceApp, "GET", "/api/rbac/check"),
		Self(AudienceConsole, "GET", "/api/rbac/menus"),
		Self(AudienceConsole, "GET", "/api/rbac/roles"),
		Self(AudienceConsole, "GET", "/api/rbac/check"),
		Need(AudienceConsole, "GET", "/api/rbac/actions", "role", "READ"),
		Need(AudienceConsole, "GET", "/api/rbac/roles/catalog", "role", "READ"),
		Need(AudienceConsole, "POST", "/api/rbac/roles", "role", "CREATE"),
		Need(AudienceConsole, "GET", "/api/rbac/roles/:role_id", "role", "READ"),
		Need(AudienceConsole, "PATCH", "/api/rbac/roles/:role_id", "role", "UPDATE"),
		Need(AudienceConsole, "GET", "/api/rbac/roles/:role_id/menus", "role", "READ"),
		Need(AudienceConsole, "PATCH", "/api/rbac/roles/:role_id/status", "role", "UPDATE"),
		Need(AudienceConsole, "PUT", "/api/rbac/roles/menus", "role", "UPDATE"),
		Need(AudienceConsole, "DELETE", "/api/rbac/roles/menus", "role", "UPDATE"),
		Need(AudienceConsole, "GET", "/api/rbac/menus/catalog", "menu", "READ"),
		Need(AudienceConsole, "POST", "/api/rbac/menus", "menu", "CREATE"),
		Need(AudienceConsole, "PATCH", "/api/rbac/menus/:menu_id", "menu", "UPDATE"),
		Need(AudienceConsole, "PATCH", "/api/rbac/menus/:menu_id/status", "menu", "UPDATE"),
		Need(AudienceConsole, "POST", "/api/rbac/users/roles", "user", "UPDATE"),
		Need(AudienceConsole, "PUT", "/api/rbac/users/roles", "user", "UPDATE"),
		Need(AudienceConsole, "DELETE", "/api/rbac/users/roles", "user", "UPDATE"),
		Need(AudienceConsole, "GET", "/api/rbac/users/:user_id/roles", "user", "READ"),
	}
}

func UserRoutes() []Route {
	return []Route{
		Self(AudienceApp, "GET", "/api/"),
		Self(AudienceApp, "PUT", "/api/"),
		Self(AudienceConsole, "GET", "/api/"),
		Self(AudienceConsole, "PUT", "/api/"),
		Need(AudienceConsole, "POST", "/api/", "user", "CREATE"),
		Need(AudienceConsole, "POST", "/api/batch", "user", "CREATE"),
		Need(AudienceConsole, "GET", "/api/exists", "user", "READ"),
		Need(AudienceConsole, "GET", "/api/page", "user", "READ"),
		Need(AudienceConsole, "GET", "/api/stats", "user", "READ"),
		Need(AudienceConsole, "GET", "/api/:id", "user", "READ"),
		Need(AudienceConsole, "DELETE", "/api/:id", "user", "DELETE"),
		Need(AudienceConsole, "PUT", "/api/:id/disable", "user", "DISABLE"),
		Need(AudienceConsole, "PUT", "/api/:id/enable", "user", "ENABLE"),
	}
}

func AccountRoutes() []Route {
	return []Route{
		Self(AudienceApp, "GET", "/api/"),
		Self(AudienceApp, "POST", "/api/"),
		Self(AudienceApp, "GET", "/api/currencies"),
		Self(AudienceApp, "GET", "/api/:id"),
		Self(AudienceApp, "POST", "/api/:id/wallets"),
		Self(AudienceApp, "GET", "/api/wallet/:id"),
		Self(AudienceConsole, "GET", "/api/currencies"),
		Need(AudienceConsole, "GET", "/api/user/:id", "account", "READ"),
		Need(AudienceConsole, "GET", "/api/:id", "account", "READ"),
		Need(AudienceConsole, "POST", "/api/", "account", "CREATE"),
		Need(AudienceConsole, "POST", "/api/:id/wallets", "account", "CREATE"),
		Need(AudienceConsole, "GET", "/api/wallet/:id", "account", "READ"),
	}
}

func AuditRoutes() []Route {
	return []Route{
		Need(AudienceConsole, "GET", "/api/events/page", "audit", "READ"),
		Need(AudienceConsole, "GET", "/api/events/:event_id", "audit", "READ"),
	}
}

func StatusRoutes() []Route {
	return []Route{
		Self(AudienceApp, "GET", "/api/status"),
		Self(AudienceConsole, "GET", "/api/status"),
	}
}

func LedgerRoutes() []Route {
	return []Route{
		Self(AudienceApp, "GET", "/api/wallet/:id"),
		Need(AudienceConsole, "GET", "/api/wallet/:id", "account", "READ"),
	}
}
