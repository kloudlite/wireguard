---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{.Release.Name}}-leader-election-role
  namespace: '{{.Release.Namespace}}'
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.Release.Name}}-manager-role
rules:
- apiGroups:
  - ""
  resources:
  - events
  - namespaces
  - secrets
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers/finalizers
  verbs:
  - update
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers/status
  verbs:
  - get
  - patch
  - update

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.Release.Name}}-metrics-auth-role
rules:
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.Release.Name}}-metrics-reader
rules:
- nonResourceURLs:
  - /metrics
  verbs:
  - get

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.Release.Name}}-server-admin-role
rules:
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers
  verbs:
  - '*'
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers/status
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.Release.Name}}-server-editor-role
rules:
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers/status
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{.Release.Name}}-server-viewer-role
rules:
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - wireguard.kloudlite.github.com
  resources:
  - servers/status
  verbs:
  - get
