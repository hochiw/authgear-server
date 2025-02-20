package service

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package service

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
	GetMany(ids []string) ([]*loginid.Identity, error)
	List(userID string) ([]*loginid.Identity, error)
	GetByValue(loginIDValue string) ([]*loginid.Identity, error)
	ListByClaim(name string, value string) ([]*loginid.Identity, error)
	New(userID string, loginID loginid.Spec, options loginid.CheckerOptions) (*loginid.Identity, error)
	WithValue(iden *loginid.Identity, value string, options loginid.CheckerOptions) (*loginid.Identity, error)
	Create(i *loginid.Identity) error
	Update(i *loginid.Identity) error
	Delete(i *loginid.Identity) error
	CheckDuplicated(uniqueKey string, standardClaims map[string]string, userID string) (*loginid.Identity, error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
	GetMany(ids []string) ([]*oauth.Identity, error)
	List(userID string) ([]*oauth.Identity, error)
	GetByProviderSubject(provider config.ProviderID, subjectID string) (*oauth.Identity, error)
	GetByUserProvider(userID string, provider config.ProviderID) (*oauth.Identity, error)
	ListByClaim(name string, value string) ([]*oauth.Identity, error)
	New(
		userID string,
		provider config.ProviderID,
		subjectID string,
		profile map[string]interface{},
		claims map[string]interface{},
	) *oauth.Identity
	WithUpdate(iden *oauth.Identity, rawProfile map[string]interface{}, claims map[string]interface{}) *oauth.Identity
	Create(i *oauth.Identity) error
	Update(i *oauth.Identity) error
	Delete(i *oauth.Identity) error
	CheckDuplicated(standardClaims map[string]string, userID string) (*oauth.Identity, error)
}

type AnonymousIdentityProvider interface {
	Get(userID, id string) (*anonymous.Identity, error)
	GetMany(ids []string) ([]*anonymous.Identity, error)
	GetByKeyID(keyID string) (*anonymous.Identity, error)
	List(userID string) ([]*anonymous.Identity, error)
	ListByClaim(name string, value string) ([]*anonymous.Identity, error)
	New(userID string, keyID string, key []byte) *anonymous.Identity
	Create(i *anonymous.Identity) error
	Delete(i *anonymous.Identity) error
}

type BiometricIdentityProvider interface {
	Get(userID, id string) (*biometric.Identity, error)
	GetMany(ids []string) ([]*biometric.Identity, error)
	GetByKeyID(keyID string) (*biometric.Identity, error)
	List(userID string) ([]*biometric.Identity, error)
	ListByClaim(name string, value string) ([]*biometric.Identity, error)
	New(userID string, keyID string, key []byte, deviceInfo map[string]interface{}) *biometric.Identity
	Create(i *biometric.Identity) error
	Delete(i *biometric.Identity) error
}

type Service struct {
	Authentication        *config.AuthenticationConfig
	Identity              *config.IdentityConfig
	IdentityFeatureConfig *config.IdentityFeatureConfig
	Store                 *Store
	LoginID               LoginIDIdentityProvider
	OAuth                 OAuthIdentityProvider
	Anonymous             AnonymousIdentityProvider
	Biometric             BiometricIdentityProvider
}

func (s *Service) Get(id string) (*identity.Info, error) {
	ref, err := s.Store.GetRefByID(id)
	if err != nil {
		return nil, err
	}
	switch ref.Type {
	case model.IdentityTypeLoginID:
		l, err := s.LoginID.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(l), nil

	case model.IdentityTypeOAuth:
		o, err := s.OAuth.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return s.toIdentityInfo(o), nil

	case model.IdentityTypeAnonymous:
		a, err := s.Anonymous.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return anonymousToIdentityInfo(a), nil
	case model.IdentityTypeBiometric:
		b, err := s.Biometric.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return biometricToIdentityInfo(b), nil
	}

	panic("identity: unknown identity type " + ref.Type)
}

func (s *Service) GetMany(ids []string) ([]*identity.Info, error) {
	refs, err := s.Store.ListRefsByIDs(ids)
	if err != nil {
		return nil, err
	}

	var loginIDs, oauthIDs, anonymousIDs, biometricIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case model.IdentityTypeLoginID:
			loginIDs = append(loginIDs, ref.ID)
		case model.IdentityTypeOAuth:
			oauthIDs = append(oauthIDs, ref.ID)
		case model.IdentityTypeAnonymous:
			anonymousIDs = append(anonymousIDs, ref.ID)
		case model.IdentityTypeBiometric:
			biometricIDs = append(biometricIDs, ref.ID)
		default:
			panic("identity: unknown identity type " + ref.Type)
		}
	}

	var infos []*identity.Info

	l, err := s.LoginID.GetMany(loginIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range l {
		infos = append(infos, loginIDToIdentityInfo(i))
	}

	o, err := s.OAuth.GetMany(oauthIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range o {
		infos = append(infos, s.toIdentityInfo(i))
	}

	a, err := s.Anonymous.GetMany(anonymousIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range a {
		infos = append(infos, anonymousToIdentityInfo(i))
	}

	b, err := s.Biometric.GetMany(biometricIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range b {
		infos = append(infos, biometricToIdentityInfo(i))
	}

	return infos, nil
}

func (s *Service) getBySpec(spec *identity.Spec) (*identity.Info, error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		loginID := extractLoginIDValue(spec.Claims)
		l, err := s.LoginID.GetByValue(loginID)
		if err != nil {
			return nil, err
		} else if len(l) != 1 {
			return nil, identity.ErrIdentityNotFound
		}
		return loginIDToIdentityInfo(l[0]), nil

	case model.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(spec.Claims)
		o, err := s.OAuth.GetByProviderSubject(providerID, subjectID)
		if err != nil {
			return nil, err
		}
		return s.toIdentityInfo(o), nil

	case model.IdentityTypeAnonymous:
		keyID, _ := extractAnonymousClaims(spec.Claims)
		if keyID != "" {
			a, err := s.Anonymous.GetByKeyID(keyID)
			if err != nil {
				return nil, err
			}
			return anonymousToIdentityInfo(a), nil
		}
		// when keyID is empty, try to get the identity from user and identity id
		userID, identityID := extractExistingIDsFromAnonymousClaims(spec.Claims)
		if userID == "" {
			return nil, identity.ErrIdentityNotFound
		}
		a, err := s.Anonymous.Get(userID, identityID)
		// identity must be found with existing user and identity id
		if err != nil {
			panic(fmt.Errorf("identity: failed to fetch anonymous identity: %s, %s, %w", userID, identityID, err))
		}
		return anonymousToIdentityInfo(a), nil
	case model.IdentityTypeBiometric:
		keyID, _, _ := extractBiometricClaims(spec.Claims)
		b, err := s.Biometric.GetByKeyID(keyID)
		if err != nil {
			return nil, err
		}
		return biometricToIdentityInfo(b), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

// SearchBySpec does not return identity.ErrIdentityNotFound.
func (s *Service) SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	exactMatch, err = s.getBySpec(spec)
	// The simplest case is the exact match case.
	if err == nil {
		return
	}

	// Any error other than identity.ErrIdentityNotFound
	if err != nil && !errors.Is(err, identity.ErrIdentityNotFound) {
		return
	}

	// Do not consider identity.ErrIdentityNotFound as error.
	err = nil

	claimsToSearch := make(map[string]interface{})

	// Otherwise we have to search.
	switch spec.Type {
	case model.IdentityTypeLoginID:
		// For login ID, we treat the login ID value as email, phone_number and preferred_username.
		loginID := extractLoginIDValue(spec.Claims)
		claimsToSearch[string(model.ClaimEmail)] = loginID
		claimsToSearch[string(model.ClaimPhoneNumber)] = loginID
		claimsToSearch[string(model.ClaimPreferredUsername)] = loginID
	case model.IdentityTypeOAuth:
		if c, ok := spec.Claims[identity.IdentityClaimOAuthClaims].(map[string]interface{}); ok {
			claimsToSearch = c
		}
	default:
		break
	}

	for name, value := range claimsToSearch {
		str, ok := value.(string)
		if !ok {
			continue
		}
		switch name {
		case string(model.ClaimEmail),
			string(model.ClaimPhoneNumber),
			string(model.ClaimPreferredUsername):

			var loginIDs []*loginid.Identity
			loginIDs, err = s.LoginID.ListByClaim(name, str)
			if err != nil {
				return
			}

			for _, loginID := range loginIDs {
				otherMatches = append(otherMatches, loginIDToIdentityInfo(loginID))
			}

			var oauths []*oauth.Identity
			oauths, err = s.OAuth.ListByClaim(name, str)
			if err != nil {
				return
			}

			for _, o := range oauths {
				otherMatches = append(otherMatches, s.toIdentityInfo(o))
			}

		}
	}

	return
}

func (s *Service) ListByUser(userID string) ([]*identity.Info, error) {
	var infos []*identity.Info

	// login id
	lis, err := s.LoginID.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		infos = append(infos, loginIDToIdentityInfo(i))
	}

	// oauth
	ois, err := s.OAuth.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, s.toIdentityInfo(i))
	}

	// anonymous
	ais, err := s.Anonymous.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ais {
		infos = append(infos, anonymousToIdentityInfo(i))
	}

	// biometric
	bis, err := s.Biometric.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range bis {
		infos = append(infos, biometricToIdentityInfo(i))
	}

	return infos, nil
}

func (s *Service) Count(userID string) (uint64, error) {
	return s.Store.Count(userID)
}

func (s *Service) ListRefsByUsers(userIDs []string) ([]*model.IdentityRef, error) {
	return s.Store.ListRefsByUsers(userIDs)
}

func (s *Service) ListByClaim(name string, value string) ([]*identity.Info, error) {
	var infos []*identity.Info

	// login id
	lis, err := s.LoginID.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		infos = append(infos, loginIDToIdentityInfo(i))
	}

	// oauth
	ois, err := s.OAuth.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, s.toIdentityInfo(i))
	}

	// anonymous
	ais, err := s.Anonymous.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ais {
		infos = append(infos, anonymousToIdentityInfo(i))
	}

	// biometric
	bis, err := s.Biometric.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range bis {
		infos = append(infos, biometricToIdentityInfo(i))
	}

	return infos, nil
}

func (s *Service) New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		loginID := extractLoginIDSpec(spec.Claims)
		l, err := s.LoginID.New(userID, loginID, loginid.CheckerOptions{
			EmailByPassBlocklistAllowlist: options.LoginIDEmailByPassBlocklistAllowlist,
		})
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(l), nil
	case model.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(spec.Claims)
		rawProfile, standardClaims := extractOAuthProfile(spec.Claims)
		o := s.OAuth.New(userID, providerID, subjectID, rawProfile, standardClaims)
		return s.toIdentityInfo(o), nil
	case model.IdentityTypeAnonymous:
		keyID, key := extractAnonymousClaims(spec.Claims)
		a := s.Anonymous.New(userID, keyID, []byte(key))
		return anonymousToIdentityInfo(a), nil
	case model.IdentityTypeBiometric:
		keyID, key, deviceInfo := extractBiometricClaims(spec.Claims)
		b := s.Biometric.New(userID, keyID, []byte(key), deviceInfo)
		return biometricToIdentityInfo(b), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

func (s *Service) Create(info *identity.Info) error {
	// TODO(verification): make OAuth verified according to config.
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(info)
		if err := s.LoginID.Create(i); err != nil {
			return err
		}
		*info = *loginIDToIdentityInfo(i)

	case model.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(info)
		if err := s.OAuth.Create(i); err != nil {
			return err
		}
		*info = *s.toIdentityInfo(i)

	case model.IdentityTypeAnonymous:
		i := anonymousFromIdentityInfo(info)
		if err := s.Anonymous.Create(i); err != nil {
			return err
		}
		*info = *anonymousToIdentityInfo(i)

	case model.IdentityTypeBiometric:
		i := biometricFromIdentityInfo(info)
		if err := s.Biometric.Create(i); err != nil {
			return err
		}
		*info = *biometricToIdentityInfo(i)

	default:
		panic("identity: unknown identity type " + info.Type)
	}
	return nil
}

func (s *Service) UpdateWithSpec(info *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i, err := s.LoginID.WithValue(loginIDFromIdentityInfo(info), extractLoginIDValue(spec.Claims), loginid.CheckerOptions{
			EmailByPassBlocklistAllowlist: options.LoginIDEmailByPassBlocklistAllowlist,
		})
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(i), nil
	case model.IdentityTypeOAuth:
		rawProfile, standardClaims := extractOAuthProfile(spec.Claims)
		i := s.OAuth.WithUpdate(
			oauthFromIdentityInfo(info),
			rawProfile,
			standardClaims,
		)
		return s.toIdentityInfo(i), nil
	default:
		panic("identity: cannot update identity type " + info.Type)
	}
}

func (s *Service) Update(info *identity.Info) error {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(info)
		if err := s.LoginID.Update(i); err != nil {
			return err
		}
		*info = *loginIDToIdentityInfo(i)

	case model.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(info)
		if err := s.OAuth.Update(i); err != nil {
			return err
		}
		*info = *s.toIdentityInfo(i)

	case model.IdentityTypeAnonymous:
		panic("identity: update no support for identity type " + info.Type)
	case model.IdentityTypeBiometric:
		panic("identity: update no support for identity type " + info.Type)
	default:
		panic("identity: unknown identity type " + info.Type)
	}

	return nil
}

func (s *Service) Delete(info *identity.Info) error {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(info)
		if err := s.LoginID.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(info)
		if err := s.OAuth.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeAnonymous:
		i := anonymousFromIdentityInfo(info)
		if err := s.Anonymous.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeBiometric:
		i := biometricFromIdentityInfo(info)
		if err := s.Biometric.Delete(i); err != nil {
			return err
		}
	default:
		panic("identity: unknown identity type " + info.Type)
	}

	return nil
}

func (s *Service) CheckDuplicated(is *identity.Info) (dupeIdentity *identity.Info, err error) {
	// extract login id unique key
	loginIDUniqueKey := ""
	if is.Type == model.IdentityTypeLoginID {
		li := loginIDFromIdentityInfo(is)
		loginIDUniqueKey = li.UniqueKey
	}

	// extract standard claims
	claims := extractStandardClaims(is.Claims)

	li, err := s.LoginID.CheckDuplicated(loginIDUniqueKey, claims, is.UserID)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		dupeIdentity = loginIDToIdentityInfo(li)
		return
	} else if err != nil {
		return
	}

	oi, err := s.OAuth.CheckDuplicated(claims, is.UserID)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		dupeIdentity = s.toIdentityInfo(oi)
		return
	} else if err != nil {
		return
	}

	// No need to consider anonymous identity

	// No need to consider biometric identity

	return
}

func (s *Service) ListCandidates(userID string) (out []identity.Candidate, err error) {
	var loginIDs []*loginid.Identity
	var oauths []*oauth.Identity

	if userID != "" {
		loginIDs, err = s.LoginID.List(userID)
		if err != nil {
			return
		}
		oauths, err = s.OAuth.List(userID)
		if err != nil {
			return
		}
		// No need to consider anonymous identity
		// No need to consider biometric identity
	}

	for _, i := range s.Authentication.Identities {
		switch i {
		case model.IdentityTypeOAuth:
			for _, providerConfig := range s.Identity.OAuth.Providers {
				pc := providerConfig
				if identity.IsOAuthSSOProviderTypeDisabled(pc.Type, s.IdentityFeatureConfig.OAuth.Providers) {
					continue
				}
				configProviderID := pc.ProviderID()
				candidate := identity.NewOAuthCandidate(&pc)
				matched := false
				for _, iden := range oauths {
					if iden.ProviderID.Equal(&configProviderID) {
						matched = true
						candidate[identity.CandidateKeyIdentityID] = iden.ID
						candidate[identity.CandidateKeyProviderSubjectID] = string(iden.ProviderSubjectID)
						candidate[identity.CandidateKeyDisplayID] = s.toIdentityInfo(iden).DisplayID()
					}
				}
				canAppend := true
				if *providerConfig.ModifyDisabled && !matched {
					canAppend = false
				}
				if canAppend {
					out = append(out, candidate)
				}
			}
		case model.IdentityTypeLoginID:
			for _, loginIDKeyConfig := range s.Identity.LoginID.Keys {
				lkc := loginIDKeyConfig
				candidate := identity.NewLoginIDCandidate(&lkc)
				matched := false
				for _, iden := range loginIDs {
					if loginIDKeyConfig.Key == iden.LoginIDKey {
						matched = true
						candidate[identity.CandidateKeyIdentityID] = iden.ID
						candidate[identity.CandidateKeyLoginIDValue] = iden.LoginID
						candidate[identity.CandidateKeyDisplayID] = loginIDToIdentityInfo(iden).DisplayID()
					}
				}
				canAppend := true
				if *loginIDKeyConfig.ModifyDisabled && !matched {
					canAppend = false
				}
				if canAppend {
					out = append(out, candidate)
				}
			}
		}
	}

	return
}

func (s *Service) toIdentityInfo(o *oauth.Identity) *identity.Info {
	provider := map[string]interface{}{
		"type": o.ProviderID.Type,
	}
	for k, v := range o.ProviderID.Keys {
		provider[k] = v
	}

	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: provider,
		identity.IdentityClaimOAuthProviderType: o.ProviderID.Type,
		identity.IdentityClaimOAuthSubjectID:    o.ProviderSubjectID,
		identity.IdentityClaimOAuthProfile:      o.UserProfile,
	}

	alias := ""
	for _, providerConfig := range s.Identity.OAuth.Providers {
		providerID := providerConfig.ProviderID()
		if providerID.Equal(&o.ProviderID) {
			alias = providerConfig.Alias
		}
	}
	if alias != "" {
		claims[identity.IdentityClaimOAuthProviderAlias] = alias
	}

	for k, v := range o.Claims {
		claims[k] = v
	}

	return &identity.Info{
		ID:        o.ID,
		UserID:    o.UserID,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Type:      model.IdentityTypeOAuth,
		Claims:    claims,
	}
}

func extractLoginIDValue(claims map[string]interface{}) string {
	loginID, ok := claims[identity.IdentityClaimLoginIDValue].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID value, got %T", claims[identity.IdentityClaimLoginIDValue]))
	}

	return loginID
}

func extractLoginIDSpec(claims map[string]interface{}) loginid.Spec {
	loginIDKey, ok := claims[identity.IdentityClaimLoginIDKey].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID key, got %T", claims[identity.IdentityClaimLoginIDKey]))
	}

	loginIDType, ok := claims[identity.IdentityClaimLoginIDType].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID type, got %T", claims[identity.IdentityClaimLoginIDType]))
	}

	loginIDValue, ok := claims[identity.IdentityClaimLoginIDValue].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID value, got %T", claims[identity.IdentityClaimLoginIDValue]))
	}

	return loginid.Spec{
		Key:   loginIDKey,
		Type:  config.LoginIDKeyType(loginIDType),
		Value: loginIDValue,
	}
}

func extractOAuthClaims(claims map[string]interface{}) (providerID config.ProviderID, subjectID string) {
	providerID = extractOAuthProviderClaims(claims)

	subjectID, ok := claims[identity.IdentityClaimOAuthSubjectID].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string subject ID claim, got %T", claims[identity.IdentityClaimOAuthSubjectID]))
	}

	return
}

func extractOAuthProfile(claims map[string]interface{}) (rawProfile map[string]interface{}, standardClaims map[string]interface{}) {
	var ok bool
	if rawProfile, ok = claims[identity.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
		rawProfile = make(map[string]interface{})
	}
	if standardClaims, ok = claims[identity.IdentityClaimOAuthClaims].(map[string]interface{}); !ok {
		standardClaims = make(map[string]interface{})
	}
	return
}

func extractOAuthProviderClaims(claims map[string]interface{}) config.ProviderID {
	provider, ok := claims[identity.IdentityClaimOAuthProviderKeys].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("identity: expect map provider claim, got %T", claims[identity.IdentityClaimOAuthProviderKeys]))
	}

	providerID := config.ProviderID{Keys: map[string]interface{}{}}
	for k, v := range provider {
		if k == "type" {
			providerID.Type, ok = v.(string)
			if !ok {
				panic(fmt.Sprintf("identity: expect string provider type, got %T", v))
			}
		} else {
			providerID.Keys[k] = v.(string)
		}
	}

	return providerID
}

func extractAnonymousClaims(claims map[string]interface{}) (keyID string, key string) {
	if v, ok := claims[identity.IdentityClaimAnonymousKeyID]; ok {
		if keyID, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string key ID, got %T", claims[identity.IdentityClaimAnonymousKeyID]))
		}
	}
	if v, ok := claims[identity.IdentityClaimAnonymousKey]; ok {
		if key, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string key, got %T", claims[identity.IdentityClaimAnonymousKey]))
		}
	}
	return
}

func extractExistingIDsFromAnonymousClaims(claims map[string]interface{}) (existingUserID string, existingIdentityID string) {
	if v, ok := claims[identity.IdentityClaimAnonymousExistingUserID]; ok {
		if existingUserID, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string existing user id, got %T", claims[identity.IdentityClaimAnonymousExistingUserID]))
		}
	}
	if v, ok := claims[identity.IdentityClaimAnonymousExistingIdentityID]; ok {
		if existingIdentityID, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string existing identity id, got %T", claims[identity.IdentityClaimAnonymousExistingIdentityID]))
		}
	}
	return
}

func extractBiometricClaims(claims map[string]interface{}) (keyID string, key string, deviceInfo map[string]interface{}) {
	if v, ok := claims[identity.IdentityClaimBiometricKeyID]; ok {
		if keyID, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string key ID, got %T", claims[identity.IdentityClaimBiometricKeyID]))
		}
	}
	if v, ok := claims[identity.IdentityClaimBiometricKey]; ok {
		if key, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string key, got %T", claims[identity.IdentityClaimBiometricKey]))
		}
	}
	if v, ok := claims[identity.IdentityClaimBiometricDeviceInfo]; ok {
		if deviceInfo, ok = v.(map[string]interface{}); !ok {
			panic(fmt.Sprintf("identity: expect map device info, got %T", claims[identity.IdentityClaimBiometricDeviceInfo]))
		}
	}
	return
}

func extractStandardClaims(claims map[string]interface{}) map[string]string {
	standardClaims := map[string]string{}
	email, hasEmail := claims[identity.StandardClaimEmail].(string)
	if hasEmail {
		standardClaims[identity.StandardClaimEmail] = email
	}

	return standardClaims
}
