import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { produce } from "immer";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ManageLanguageWidget from "./ManageLanguageWidget";
import ImageFilePicker from "../../ImageFilePicker";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { useTemplateLocaleQuery } from "./query/templateLocaleQuery";
import { PortalAPIAppConfig } from "../../types";
import {
  ALL_LOCALIZABLE_RESOURCES,
  ALL_TEMPLATES,
  renderPath,
  RESOURCE_APP_BANNER,
  RESOURCE_APP_LOGO,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT,
  RESOURCE_FORGOT_PASSWORD_EMAIL_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT,
  RESOURCE_TRANSLATION_JSON,
} from "../../resources";
import {
  LanguageTag,
  LocaleInvalidReason,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
  validateLocales,
} from "../../util/resource";

import styles from "./ResourceConfigurationScreen.module.scss";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { clearEmptyObject } from "../../util/misc";
import { useResourceForm } from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

interface ConfigFormState {
  defaultLocale: string;
}

const defaultConfigFormState: ConfigFormState = { defaultLocale: "en" };

function constructConfigFormState(config: PortalAPIAppConfig): ConfigFormState {
  return { defaultLocale: config.localization?.fallback_language ?? "en" };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: ConfigFormState,
  currentState: ConfigFormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.localization = config.localization ?? {};
    if (initialState.defaultLocale !== currentState.defaultLocale) {
      config.localization.fallback_language = currentState.defaultLocale;
    }
    clearEmptyObject(config);
  });
}

interface ResourcesFormState {
  resources: Partial<Record<string, Resource>>;
}

function constructResourcesFormState(
  resources: Resource[]
): ResourcesFormState {
  const resourceMap: Record<string, Resource> = {};
  for (const r of resources) {
    resourceMap[specifierId(r.specifier)] = r;
  }
  return { resources: resourceMap };
}

function constructResources(state: ResourcesFormState): Resource[] {
  return Object.values(state.resources).filter(Boolean) as Resource[];
}

interface FormState extends ConfigFormState, ResourcesFormState {
  selectedLocale: string;
}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => void;
}

interface ResourcesConfigurationContentProps {
  form: FormModel;
  locales: LanguageTag[];
  invalidLocaleReason: LocaleInvalidReason | null;
  invalidLocales: LanguageTag[];
}

const PIVOT_KEY_APPEARANCE = "appearance";
const PIVOT_KEY_FORGOT_PASSWORD = "forgot_password";
const PIVOT_KEY_PASSWORDLESS = "passwordless";
const PIVOT_KEY_TRANSLATION_JSON = "translation.json";

const ALL_PIVOT_KEYS = [
  PIVOT_KEY_TRANSLATION_JSON,
  PIVOT_KEY_APPEARANCE,
  PIVOT_KEY_FORGOT_PASSWORD,
  PIVOT_KEY_PASSWORDLESS,
];

const ResourcesConfigurationContent: React.FC<ResourcesConfigurationContentProps> = function ResourcesConfigurationContent(
  props
) {
  const { state, setState } = props.form;
  const { locales, invalidLocaleReason, invalidLocales } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="ResourceConfigurationScreen.title" />,
      },
    ];
  }, []);

  const setDefaultLocale = useCallback(
    (defaultLocale: LanguageTag) => {
      setState((s) => ({ ...s, defaultLocale }));
    },
    [setState]
  );

  const setSelectedLocale = useCallback(
    (selectedLocale: LanguageTag) => {
      setState((s) => ({ ...s, selectedLocale }));
    },
    [setState]
  );

  const onChangeLocales = useCallback(
    (newLocales: LanguageTag[]) => {
      setState((prev) => {
        let { selectedLocale, resources } = prev;
        resources = { ...resources };

        // Reset selected locale to default if it's removed.
        if (!newLocales.includes(state.selectedLocale)) {
          selectedLocale = state.defaultLocale;
        }

        // Populate initial resources for added locales from default locale.
        const addedLocales = newLocales.filter((l) => !locales.includes(l));
        for (const locale of addedLocales) {
          for (const resource of ALL_TEMPLATES) {
            const defaultResource =
              state.resources[
                specifierId({ def: resource, locale: prev.defaultLocale })
              ];
            const newResource: Resource = {
              specifier: {
                def: resource,
                locale,
              },
              path: renderPath(resource.resourcePath, { locale }),
              value: defaultResource?.value ?? "",
            };
            resources[specifierId(newResource.specifier)] = newResource;
          }
        }

        // Remove resources of removed locales.
        const removedLocales = locales.filter((l) => !newLocales.includes(l));
        for (const [id, resource] of Object.entries(resources)) {
          const locale = resource?.specifier.locale;
          if (resource && locale && removedLocales.includes(locale)) {
            resources[id] = { ...resource, value: "" };
          }
        }

        return { ...prev, selectedLocale, resources };
      });
    },
    [locales, setState, state]
  );

  const [selectedKey, setSelectedKey] = useState<string>(
    PIVOT_KEY_TRANSLATION_JSON
  );
  const onLinkClick = useCallback((item?: PivotItem) => {
    const itemKey = item?.props.itemKey;
    if (itemKey != null) {
      const idx = ALL_PIVOT_KEYS.indexOf(itemKey);
      if (idx >= 0) {
        setSelectedKey(itemKey);
      }
    }
  }, []);

  const getValueIgnoreEmptyString = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      const resource = state.resources[specifierId(specifier)];
      if (resource == null || resource.value === "") {
        return undefined;
      }
      return resource.value;
    },
    [state.resources, state.selectedLocale]
  );

  const getValue = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      const resource = state.resources[specifierId(specifier)];
      return resource?.value ?? "";
    },
    [state.resources, state.selectedLocale]
  );

  const getOnChange = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      return (_e: unknown, value?: string) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const resource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {
              locale: specifier.locale,
            }),
            value: value ?? "",
          };
          updatedResources[specifierId(resource.specifier)] = resource;
          return { ...prev, resources: updatedResources };
        });
      };
    },
    [state.selectedLocale, setState]
  );

  const getOnChangeImage = useCallback(
    (def: ResourceDefinition) => {
      const specifier: ResourceSpecifier = {
        def,
        locale: state.selectedLocale,
      };
      return (base64EncodedData?: string, extension?: string) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const resource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {
              locale: specifier.locale,
              extension,
            }),
            value: base64EncodedData ?? "",
          };
          updatedResources[specifierId(resource.specifier)] = resource;
          return { ...prev, resources: updatedResources };
        });
      };
    },
    [state.selectedLocale, setState]
  );

  const sectionsTranslationJSON: EditTemplatesWidgetSection[] = [
    {
      key: "translation.json",
      title: (
        <FormattedMessage id="EditTemplatesWidget.translationjson.title" />
      ),
      items: [
        {
          key: "translation.json",
          title: (
            <FormattedMessage id="EditTemplatesWidget.translationjson.subtitle" />
          ),
          language: "json",
          value: getValue(RESOURCE_TRANSLATION_JSON),
          onChange: getOnChange(RESOURCE_TRANSLATION_JSON),
        },
      ],
    },
  ];

  const sectionsForgotPassword: EditTemplatesWidgetSection[] = [
    {
      key: "email",
      title: <FormattedMessage id="EditTemplatesWidget.email" />,
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
        },
      ],
    },
    {
      key: "sms",
      title: <FormattedMessage id="EditTemplatesWidget.sms" />,
      items: [
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_FORGOT_PASSWORD_SMS_TXT),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_SMS_TXT),
        },
      ],
    },
  ];

  const sectionsPasswordless: EditTemplatesWidgetSection[] = [
    {
      key: "setup",
      title: (
        <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
      ),
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
        },
      ],
    },
    {
      key: "login",
      title: (
        <FormattedMessage id="EditTemplatesWidget.passwordless.login.title" />
      ),
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
        },
      ],
    },
  ];

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <ManageLanguageWidget
        templateLocales={locales}
        onChangeTemplateLocales={onChangeLocales}
        templateLocale={state.selectedLocale}
        defaultTemplateLocale={state.defaultLocale}
        onSelectTemplateLocale={setSelectedLocale}
        onSelectDefaultTemplateLocale={setDefaultLocale}
        invalidLocaleReason={invalidLocaleReason}
        invalidTemplateLocales={invalidLocales}
      />
      <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.translationjson.title"
          )}
          itemKey={PIVOT_KEY_TRANSLATION_JSON}
        >
          <EditTemplatesWidget sections={sectionsTranslationJSON} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.appearance.title"
          )}
          itemKey={PIVOT_KEY_APPEARANCE}
        >
          <div className={styles.pivotItemAppearance}>
            <ImageFilePicker
              title={renderToString("ResourceConfigurationScreen.app-banner")}
              base64EncodedData={getValueIgnoreEmptyString(RESOURCE_APP_BANNER)}
              onChange={getOnChangeImage(RESOURCE_APP_BANNER)}
            />
            <ImageFilePicker
              title={renderToString("ResourceConfigurationScreen.app-logo")}
              base64EncodedData={getValueIgnoreEmptyString(RESOURCE_APP_LOGO)}
              onChange={getOnChangeImage(RESOURCE_APP_LOGO)}
            />
          </div>
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.forgot-password.title"
          )}
          itemKey={PIVOT_KEY_FORGOT_PASSWORD}
        >
          <EditTemplatesWidget sections={sectionsForgotPassword} />
        </PivotItem>
        <PivotItem
          headerText={renderToString(
            "ResourceConfigurationScreen.passwordless-authenticator.title"
          )}
          itemKey={PIVOT_KEY_PASSWORDLESS}
        >
          <EditTemplatesWidget sections={sectionsPasswordless} />
        </PivotItem>
      </Pivot>
    </div>
  );
};

const ResourceConfigurationScreen: React.FC = function ResourceConfigurationScreen() {
  const { appID } = useParams();

  const {
    templateLocales: resourceLocales,
    loading: isLoadingLocales,
    error: loadLocalesError,
    refetch: reloadLocales,
  } = useTemplateLocaleQuery(appID);

  const specifiers = useMemo<ResourceSpecifier[]>(() => {
    const specifiers = [];
    for (const locale of resourceLocales) {
      for (const def of ALL_LOCALIZABLE_RESOURCES) {
        specifiers.push({
          def,
          locale,
        });
      }
    }
    return specifiers;
  }, [resourceLocales]);

  const [selectedLocale, setSelectedLocale] = useState<LanguageTag | null>(
    null
  );

  const config = useAppConfigForm(
    appID,
    defaultConfigFormState,
    constructConfigFormState,
    constructConfig
  );
  const resources = useResourceForm(
    appID,
    specifiers,
    constructResourcesFormState,
    constructResources
  );
  const state = useMemo<FormState>(
    () => ({
      defaultLocale: config.state.defaultLocale,
      resources: resources.state.resources,
      selectedLocale: selectedLocale ?? config.state.defaultLocale,
    }),
    [config.state.defaultLocale, resources.state.resources, selectedLocale]
  );

  const form: FormModel = {
    isLoading: isLoadingLocales || config.isLoading || resources.isLoading,
    isUpdating: config.isUpdating || resources.isUpdating,
    isDirty: config.isDirty || resources.isDirty,
    loadError: loadLocalesError ?? config.loadError ?? resources.loadError,
    updateError: config.updateError ?? resources.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      config.setState(() => ({ defaultLocale: newState.defaultLocale }));
      resources.setState(() => ({ resources: newState.resources }));
      setSelectedLocale(newState.selectedLocale);
    },
    reload: () => {
      reloadLocales().catch(() => {});
      config.reload();
      resources.reload();
    },
    reset: () => {
      config.reset();
      resources.reset();
      setSelectedLocale(config.state.defaultLocale);
    },
    save: () => {
      config.save();
      resources.save();
    },
  };

  const locales = useMemo<LanguageTag[]>(() => {
    const locales = new Set<LanguageTag>();
    for (const r of Object.values(resources.state.resources)) {
      if (r?.specifier.locale && r.value.length !== 0) {
        locales.add(r.specifier.locale);
      }
    }
    if (selectedLocale) {
      locales.add(selectedLocale);
    }
    locales.add(config.state.defaultLocale);
    // eslint-disable-next-line @typescript-eslint/require-array-sort-compare
    return Array.from(locales).sort();
  }, [selectedLocale, resources.state, config.state]);

  const { invalidReason: invalidLocaleReason, invalidLocales } = useMemo(
    () =>
      validateLocales(
        config.state.defaultLocale,
        locales,
        Object.values(resources.state.resources).filter(Boolean) as Resource[]
      ),
    [locales, config.state, resources.state]
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form} canSave={invalidLocaleReason == null}>
      <ResourcesConfigurationContent
        form={form}
        locales={locales}
        invalidLocaleReason={invalidLocaleReason}
        invalidLocales={invalidLocales}
      />
    </FormContainer>
  );
};

export default ResourceConfigurationScreen;
