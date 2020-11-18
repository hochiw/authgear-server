/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AppTemplatesQuery
// ====================================================

export interface AppTemplatesQuery_node_User {
  __typename: "User";
}

export interface AppTemplatesQuery_node_App_resources {
  __typename: "AppResource";
  path: string;
  effectiveData: string | null;
}

export interface AppTemplatesQuery_node_App_resourceLocales {
  __typename: "AppResource";
  path: string;
  languageTag: string | null;
}

export interface AppTemplatesQuery_node_App {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  resources: AppTemplatesQuery_node_App_resources[];
  resourceLocales: AppTemplatesQuery_node_App_resourceLocales[];
}

export type AppTemplatesQuery_node = AppTemplatesQuery_node_User | AppTemplatesQuery_node_App;

export interface AppTemplatesQuery {
  /**
   * Fetches an object given its ID
   */
  node: AppTemplatesQuery_node | null;
}

export interface AppTemplatesQueryVariables {
  id: string;
  paths: string[];
}
