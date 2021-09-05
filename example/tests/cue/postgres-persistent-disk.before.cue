instance_groups: [...#ig]

#ig: #inscope | #notinscope

#inscope: {
	name:                 "postgres_autoscaler"
	persistent_disk_type: null
	...
}

#notinscope: {
	name: !="postgres_autoscaler"
	...
}
