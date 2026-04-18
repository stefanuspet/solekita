package model

type Permission string

const (
	PermissionManageOrder    Permission = "manage_order"
	PermissionUpdateStatus   Permission = "update_status"
	PermissionManageDelivery Permission = "manage_delivery"
	PermissionViewReport     Permission = "view_report"
	PermissionManageOutlet   Permission = "manage_outlet"
)
