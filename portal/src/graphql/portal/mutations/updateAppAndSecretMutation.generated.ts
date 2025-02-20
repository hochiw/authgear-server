import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateAppAndSecretConfigMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID'];
  appConfig: Types.Scalars['AppConfig'];
  secretConfig?: Types.InputMaybe<Types.SecretConfigInput>;
}>;


export type UpdateAppAndSecretConfigMutationMutation = { __typename?: 'Mutation', updateApp: { __typename?: 'UpdateAppPayload', app: { __typename?: 'App', id: string, effectiveAppConfig: any, rawAppConfig: any, secretConfig: { __typename?: 'SecretConfig', oauthClientSecrets?: Array<{ __typename?: 'OAuthClientSecret', alias: string, clientSecret: string }> | null, webhookSecret?: { __typename?: 'WebhookSecret', secret?: string | null } | null, adminAPISecrets?: Array<{ __typename?: 'AdminAPISecret', keyID: string, createdAt?: any | null, publicKeyPEM: string, privateKeyPEM?: string | null }> | null, smtpSecret?: { __typename?: 'SMTPSecret', host: string, port: number, username: string, password?: string | null } | null } } } };


export const UpdateAppAndSecretConfigMutationDocument = gql`
    mutation updateAppAndSecretConfigMutation($appID: ID!, $appConfig: AppConfig!, $secretConfig: SecretConfigInput) {
  updateApp(
    input: {appID: $appID, appConfig: $appConfig, secretConfig: $secretConfig}
  ) {
    app {
      id
      effectiveAppConfig
      rawAppConfig
      secretConfig {
        oauthClientSecrets {
          alias
          clientSecret
        }
        webhookSecret {
          secret
        }
        adminAPISecrets {
          keyID
          createdAt
          publicKeyPEM
          privateKeyPEM
        }
        smtpSecret {
          host
          port
          username
          password
        }
      }
    }
  }
}
    `;
export type UpdateAppAndSecretConfigMutationMutationFn = Apollo.MutationFunction<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>;

/**
 * __useUpdateAppAndSecretConfigMutationMutation__
 *
 * To run a mutation, you first call `useUpdateAppAndSecretConfigMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateAppAndSecretConfigMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateAppAndSecretConfigMutationMutation, { data, loading, error }] = useUpdateAppAndSecretConfigMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      appConfig: // value for 'appConfig'
 *      secretConfig: // value for 'secretConfig'
 *   },
 * });
 */
export function useUpdateAppAndSecretConfigMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>(UpdateAppAndSecretConfigMutationDocument, options);
      }
export type UpdateAppAndSecretConfigMutationMutationHookResult = ReturnType<typeof useUpdateAppAndSecretConfigMutationMutation>;
export type UpdateAppAndSecretConfigMutationMutationResult = Apollo.MutationResult<UpdateAppAndSecretConfigMutationMutation>;
export type UpdateAppAndSecretConfigMutationMutationOptions = Apollo.BaseMutationOptions<UpdateAppAndSecretConfigMutationMutation, UpdateAppAndSecretConfigMutationMutationVariables>;