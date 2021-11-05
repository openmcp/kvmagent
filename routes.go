package main

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"deletekvmnode",
		"GET",
		"/deletekvmnode",
		DeleteNode,
	},
	Route{
		"createkvmnode",
		"GET",
		"/createkvmnode",
		CreateNode,
	},
	Route{
		"changekvmnode",
		"GET",
		"/changekvmnode",
		ChangeNode,
	},
	Route{
		"getKVMLists",
		"GET",
		"/getkvmlists",
		GetKVMLists,
	},
	Route{
		"kvmstopnode",
		"GET",
		"/kvmstopnode",
		StopNode,
	},
	Route{
		"kvmstartnode",
		"GET",
		"/kvmstartnode",
		StartNode,
	},
	Route{
		"getmastervmname",
		"GET",
		"/getmastervmname",
		GetMasterVMName,
	},
}
