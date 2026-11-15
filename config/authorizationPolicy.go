package config

import (
	"fmt"
	"log"
	"slices"
	"strings"
)

type (
	Permission    string
	AccessScope   string
	PrincipalKind string
	ScopesByGrant map[Permission][]AccessScope
)

type Grant struct {
	Permission Permission
	Scope      AccessScope
}

const (
	ScopeOwn AccessScope = "own"
	ScopeAny AccessScope = "any"

	PrincipalKindUser      PrincipalKind = "user"
	PrincipalKindAPIClient PrincipalKind = "api_client"

	RoleCandidate = "CANDIDATE"
	RoleAdmin     = "ADMIN"

	grantSeparator = ":"
	grantParts     = 3
)

const (
	PermissionCandidateRead          Permission = "candidate:read"
	PermissionCandidateUpdate        Permission = "candidate:update"
	PermissionResumeCreate           Permission = "resume:create"
	PermissionResumeRead             Permission = "resume:read"
	PermissionResumeUpdate           Permission = "resume:update"
	PermissionResumeDelete           Permission = "resume:delete"
	PermissionMatchRead              Permission = "match:read"
	PermissionContactRead            Permission = "contact:read"
	PermissionContactEnrich          Permission = "contact:enrich"
	PermissionJobRead                Permission = "job:read"
	PermissionJobIngest              Permission = "job:ingest"
	PermissionAPIClientCreate        Permission = "api_client:create"
	PermissionAPIClientRead          Permission = "api_client:read"
	PermissionAPIClientRevoke        Permission = "api_client:revoke"
	PermissionJobSourceDelete        Permission = "job_source:delete"
	PermissionJobSourceUpdate        Permission = "job_source:update"
	PermissionJobSourceRead          Permission = "job_source:read"
	PermissionJobSourceCreate        Permission = "job_source:create"
	PermissionChannelAccountCreate   Permission = "channel_account:create"
	PermissionChannelAccountRead     Permission = "channel_account:read"
	PermissionChannelAccountUpdate   Permission = "channel_account:update"
	PermissionChannelAccountDelete   Permission = "channel_account:delete"
	PermissionOutreachTemplateCreate Permission = "outreach_template:create"
	PermissionOutreachTemplateRead   Permission = "outreach_template:read"
	PermissionOutreachTemplateUpdate Permission = "outreach_template:update"
	PermissionOutreachTemplateDelete Permission = "outreach_template:delete"
	PermissionOutreachCampaignCreate Permission = "outreach_campaign:create"
	PermissionOutreachCampaignRead   Permission = "outreach_campaign:read"
	PermissionOutreachMessageCreate  Permission = "outreach_message:create"
	PermissionOutreachMessageRead    Permission = "outreach_message:read"
	PermissionOutreachMessageUpdate  Permission = "outreach_message:update"
	PermissionApplicationUpdate      Permission = "application:update"
	PermissionApplicationRead        Permission = "application:read"
	PermissionApplicationCreate      Permission = "application:create"
	PermissionOutreachCampaignUpdate Permission = "outreach_campaign:update"
)

var permissionCatalogueI = []Permission{
	PermissionCandidateRead,
	PermissionCandidateUpdate,
	PermissionResumeCreate,
	PermissionResumeRead,
	PermissionResumeUpdate,
	PermissionResumeDelete,
	PermissionMatchRead,
	PermissionContactRead,
	PermissionContactEnrich,
	PermissionJobRead,
	PermissionJobIngest,
	PermissionAPIClientCreate,
	PermissionAPIClientRead,
	PermissionAPIClientRevoke,
	PermissionJobSourceDelete,
	PermissionJobSourceUpdate,
	PermissionJobSourceRead,
	PermissionJobSourceCreate,
	PermissionChannelAccountCreate,
	PermissionChannelAccountRead,
	PermissionChannelAccountUpdate,
	PermissionChannelAccountDelete,
	PermissionOutreachTemplateCreate,
	PermissionOutreachTemplateRead,
	PermissionOutreachTemplateUpdate,
	PermissionOutreachTemplateDelete,
	PermissionOutreachCampaignCreate,
	PermissionOutreachCampaignRead,
	PermissionOutreachCampaignUpdate,
	PermissionOutreachMessageUpdate,
	PermissionApplicationUpdate,
	PermissionApplicationRead,
	PermissionApplicationCreate,
	PermissionOutreachMessageRead,
	PermissionOutreachMessageCreate,
}

var accessScopesI = []AccessScope{ScopeOwn, ScopeAny}

var rolePoliciesI = map[string][]string{
	RoleCandidate: {
		"candidate:read:own",
		"candidate:update:own",
		"resume:create:own",
		"resume:read:own",
		"resume:update:own",
		"resume:delete:own",
		"match:read:own",
		"contact:read:own",
		"contact:enrich:own",
		"job:read:any",
		"channel_account:create:own",
		"channel_account:read:own",
		"channel_account:update:own",
		"channel_account:delete:own",
		"outreach_template:create:own",
		"outreach_template:read:own",
		"outreach_template:update:own",
		"outreach_template:delete:own",
		"outreach_campaign:create:own",
		"outreach_campaign:read:own",
		"outreach_campaign:update:own",
		"outreach_message:update:own",
		"application:update:own",
		"application:read:own",
		"application:create:own",
		"outreach_message:read:own",
		"outreach_message:create:own",
	},
	RoleAdmin: grantEveryPermission(ScopeAny),
}

var (
	defaultSignupRolesI = []string{RoleCandidate}
	apiClientGrantsI    = []string{"job:ingest:any"}
)

type authorizationPolicy struct {
	roleGrants      map[string][]Grant
	apiClientGrants map[string]Grant
}

var authorizationPolicyI authorizationPolicy

func loadAuthorizationPolicy() {
	policy, err := buildAuthorizationPolicy()
	if err != nil {
		log.Fatalf("invalid authorization policy: %v", err)
	}

	authorizationPolicyI = policy
}

func buildAuthorizationPolicy() (authorizationPolicy, error) {
	policy := authorizationPolicy{
		roleGrants:      make(map[string][]Grant, len(rolePoliciesI)),
		apiClientGrants: make(map[string]Grant, len(apiClientGrantsI)),
	}

	for role, rawGrants := range rolePoliciesI {
		for _, rawGrant := range rawGrants {
			grant, err := ParseGrant(rawGrant)
			if err != nil {
				return authorizationPolicy{}, fmt.Errorf("role %s: %w", role, err)
			}

			policy.roleGrants[role] = append(policy.roleGrants[role], grant)
		}
	}

	for _, role := range defaultSignupRolesI {
		if _, found := policy.roleGrants[role]; !found {
			return authorizationPolicy{}, fmt.Errorf("default role %s: %w", role, ErrUnknownRole)
		}
	}

	for _, rawGrant := range apiClientGrantsI {
		grant, err := ParseGrant(rawGrant)
		if err != nil {
			return authorizationPolicy{}, fmt.Errorf("api client grants: %w", err)
		}

		policy.apiClientGrants[rawGrant] = grant
	}

	return policy, nil
}

func grantEveryPermission(scope AccessScope) []string {
	grants := make([]string, 0, len(permissionCatalogueI))
	for _, permission := range permissionCatalogueI {
		grants = append(grants, string(permission)+grantSeparator+string(scope))
	}

	return grants
}

func ParseGrant(rawGrant string) (Grant, error) {
	parts := strings.Split(rawGrant, grantSeparator)
	if len(parts) != grantParts {
		return Grant{}, fmt.Errorf("%w: %q", ErrInvalidGrant, rawGrant)
	}

	grant := Grant{
		Permission: Permission(parts[0] + grantSeparator + parts[1]),
		Scope:      AccessScope(parts[2]),
	}

	if !slices.Contains(permissionCatalogueI, grant.Permission) ||
		!slices.Contains(accessScopesI, grant.Scope) {
		return Grant{}, fmt.Errorf("%w: %q", ErrInvalidGrant, rawGrant)
	}

	return grant, nil
}

func ResolveRoleGrants(roles []string) ScopesByGrant {
	resolved := ScopesByGrant{}

	for _, role := range roles {
		resolved.add(authorizationPolicyI.roleGrants[role]...)
	}

	return resolved
}

func ResolveAPIClientGrants(rawGrants []string) ScopesByGrant {
	resolved := ScopesByGrant{}

	for _, rawGrant := range rawGrants {
		if grant, found := authorizationPolicyI.apiClientGrants[rawGrant]; found {
			resolved.add(grant)
		}
	}

	return resolved
}

func (s ScopesByGrant) add(grants ...Grant) {
	for _, grant := range grants {
		if !slices.Contains(s[grant.Permission], grant.Scope) {
			s[grant.Permission] = append(s[grant.Permission], grant.Scope)
		}
	}
}

func IsKnownRole(role string) bool {
	_, found := authorizationPolicyI.roleGrants[role]

	return found
}

func IsAPIClientGrant(rawGrant string) bool {
	_, found := authorizationPolicyI.apiClientGrants[rawGrant]

	return found
}

func GetDefaultSignupRoles() []string { return slices.Clone(defaultSignupRolesI) }
func GetPermissions() []Permission    { return slices.Clone(permissionCatalogueI) }
