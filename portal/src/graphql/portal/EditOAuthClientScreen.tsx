import cn from "classnames";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import deepEqual from "deep-equal";
import produce, { createDraft } from "immer";
import {
  Icon,
  Text,
  Link,
  useTheme,
  Image,
  ImageFit,
  Dialog,
  IDialogContentProps,
  DefaultButton,
  DialogFooter,
  ICommandBarItemProps,
  PrimaryButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import EditOAuthClientForm, {
  getReducedClientConfig,
} from "./EditOAuthClientForm";
import {
  ApplicationType,
  OAuthClientConfig,
  PortalAPIAppConfig,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import styles from "./EditOAuthClientScreen.module.css";
import Widget from "../../Widget";
import flutterIconURL from "../../images/framework_flutter.svg";
import xamarinIconURL from "../../images/framework_xamarin.svg";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useSystemConfig } from "../../context/SystemConfigContext";

interface FormState {
  publicOrigin: string;
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
  removeClientByID?: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
    clients: config.oauth?.clients ?? [],
    editedClient: null,
    removeClientByID: undefined,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients.slice();

    if (currentState.removeClientByID) {
      config.oauth.clients = config.oauth.clients.filter(
        (c) => c.client_id !== currentState.removeClientByID
      );
      clearEmptyObject(config);
      return;
    }

    const client = currentState.editedClient;
    if (client) {
      const index = config.oauth.clients.findIndex(
        (c) => c.client_id === client.client_id
      );
      if (
        index !== -1 &&
        !deepEqual(
          getReducedClientConfig(client),
          getReducedClientConfig(config.oauth.clients[index]),
          { strict: true }
        )
      ) {
        config.oauth.clients[index] = createDraft(client);
      }
    }
    clearEmptyObject(config);
  });
}

interface FrameworkItem {
  icon: React.ReactNode;
  name: string;
  docLink: string;
}

interface QuickStartFrameworkItemProps extends FrameworkItem {
  showOpenTutorialLabelWhenHover: boolean;
}

const QuickStartFrameworkItem: React.FC<QuickStartFrameworkItemProps> =
  function QuickStartFrameworkItem(props) {
    const { icon, name, docLink, showOpenTutorialLabelWhenHover } = props;
    const [isHovering, setIsHovering] = useState(false);

    const onMouseOver = useCallback(() => {
      setIsHovering(true);
    }, [setIsHovering]);

    const onMouseOut = useCallback(() => {
      setIsHovering(false);
    }, [setIsHovering]);

    // when shouldShowArrowIcon is false, open tutorial label will be shown instead
    const shouldShowArrowIcon = useMemo(() => {
      // always show open tutorial label
      if (!showOpenTutorialLabelWhenHover) {
        return true;
      }

      // show open tutorial label when hover
      return !isHovering;
    }, [showOpenTutorialLabelWhenHover, isHovering]);

    return (
      <Link
        onMouseOver={onMouseOver}
        onMouseOut={onMouseOut}
        className={cn(styles.quickStartItem, {
          [styles.quickStartItemHovered]: isHovering,
        })}
        href={docLink}
        target="_blank"
      >
        <span className={styles.quickStartItemIcon}>{icon}</span>
        <Text variant="small" className={styles.quickStartItemText}>
          {name}
        </Text>
        {shouldShowArrowIcon && (
          <Icon
            className={styles.quickStartItemArrowIcon}
            iconName="ChevronRightSmall"
          />
        )}
        {!shouldShowArrowIcon && (
          <Text className={styles.quickStartItemOpenTutorial}>
            <FormattedMessage id="EditOAuthClientScreen.quick-start.open-tutorial.label" />
          </Text>
        )}
      </Link>
    );
  };

interface QuickStartFrameworkListProps {
  applicationType?: ApplicationType;
  showOpenTutorialLabelWhenHover: boolean;
}

const QuickStartFrameworkList: React.FC<QuickStartFrameworkListProps> =
  function QuickStartFrameworkList(props) {
    const { applicationType, showOpenTutorialLabelWhenHover } = props;
    const { renderToString } = useContext(Context);

    const items: FrameworkItem[] = useMemo(() => {
      switch (applicationType) {
        case "spa":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react"
              ),
              docLink: "https://docs.authgear.com/tutorials/spa/react",
            },
            {
              icon: <i className={cn("fab", "fa-vuejs")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.vue"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
            {
              icon: <i className={cn("fab", "fa-angular")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.angular"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
            {
              icon: <i className={cn("fab", "fa-js")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.other-js"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
          ];
        case "traditional_webapp":
          return [
            {
              icon: <Icon iconName="Globe" />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.traditional-webapp"
              ),
              docLink: "https://docs.authgear.com/get-started/website",
            },
          ];
        case "native":
          return [
            {
              icon: <i className={cn("fab", "fa-react")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.react-native"
              ),
              docLink: "https://docs.authgear.com/get-started/react-native",
            },
            {
              icon: <i className={cn("fab", "fa-apple")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.ios"
              ),
              docLink: "https://docs.authgear.com/get-started/ios",
            },
            {
              icon: <i className={cn("fab", "fa-android")} />,
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.android"
              ),
              docLink: "https://docs.authgear.com/get-started/android",
            },
            {
              icon: (
                <Image
                  src={flutterIconURL}
                  imageFit={ImageFit.contain}
                  className={styles.frameworkImage}
                />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.flutter"
              ),
              docLink: "https://docs.authgear.com/get-started/flutter",
            },
            {
              icon: (
                <Image
                  src={xamarinIconURL}
                  imageFit={ImageFit.contain}
                  className={styles.frameworkImage}
                />
              ),
              name: renderToString(
                "EditOAuthClientScreen.quick-start.framework.xamarin"
              ),
              docLink: "https://docs.authgear.com/get-started/xamarin",
            },
          ];
        default:
          return [];
      }
    }, [applicationType, renderToString]);

    if (applicationType == null) {
      return null;
    }

    return (
      <>
        {items.map((item, index) => (
          <QuickStartFrameworkItem
            key={`quick-start-${index}`}
            showOpenTutorialLabelWhenHover={showOpenTutorialLabelWhenHover}
            {...item}
          />
        ))}
      </>
    );
  };

interface EditOAuthClientNavBreadcrumbProps {
  clientName: string;
}

const EditOAuthClientNavBreadcrumb: React.FC<EditOAuthClientNavBreadcrumbProps> =
  function EditOAuthClientNavBreadcrumb(props) {
    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "./../..",
          label: (
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: props.clientName,
        },
      ];
    }, [props.clientName]);

    return (
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
    );
  };

interface EditOAuthClientContentProps {
  form: AppConfigFormModel<FormState>;
  clientID: string;
}

const EditOAuthClientContent: React.FC<EditOAuthClientContentProps> =
  function EditOAuthClientContent(props) {
    const {
      clientID,
      form: { state, setState },
    } = props;
    const theme = useTheme();

    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const onClientConfigChange = useCallback(
      (editedClient: OAuthClientConfig) => {
        setState((state) => ({ ...state, editedClient }));
      },
      [setState]
    );

    if (client == null) {
      return (
        <Text>
          <FormattedMessage
            id="EditOAuthClientScreen.client-not-found"
            values={{ clientID }}
          />
        </Text>
      );
    }

    return (
      <ScreenContent>
        <EditOAuthClientNavBreadcrumb clientName={client.name ?? ""} />
        <div className={cn(styles.widget, styles.widgetColumn)}>
          <EditOAuthClientForm
            publicOrigin={state.publicOrigin}
            clientConfig={client}
            onClientConfigChange={onClientConfigChange}
          />
        </div>
        <div className={styles.quickStartColumn}>
          <Widget>
            <div className={styles.quickStartWidget}>
              <Text className={styles.quickStartWidgetTitle}>
                <Icon
                  className={styles.quickStartWidgetTitleIcon}
                  styles={{ root: { color: theme.palette.themePrimary } }}
                  iconName="Lightbulb"
                />
                <FormattedMessage id="EditOAuthClientScreen.quick-start-widget.title" />
              </Text>
              <Text>
                <FormattedMessage id="EditOAuthClientScreen.quick-start-widget.question" />
              </Text>
              <QuickStartFrameworkList
                applicationType={client.x_application_type}
                showOpenTutorialLabelWhenHover={false}
              />
            </div>
          </Widget>
        </div>
      </ScreenContent>
    );
  };

interface OAuthQuickStartScreenContentProps {
  form: AppConfigFormModel<FormState>;
  clientID: string;
}

const OAuthQuickStartScreenContent: React.FC<OAuthQuickStartScreenContentProps> =
  function OAuthQuickStartScreenContent(props) {
    const {
      clientID,
      form: { state },
    } = props;
    const navigate = useNavigate();
    const theme = useTheme();
    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const onNextButtonClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        navigate(".");
      },
      [navigate]
    );

    return (
      <ScreenLayoutScrollView>
        <ScreenContent>
          <EditOAuthClientNavBreadcrumb clientName={client?.name ?? ""} />
          <Widget className={styles.widget}>
            <Text variant="xLarge" block={true}>
              <Icon
                className={styles.quickStartScreenTitleIcon}
                styles={{ root: { color: theme.palette.themePrimary } }}
                iconName="Lightbulb"
              />
              <FormattedMessage id="EditOAuthClientScreen.quick-start-screen.title" />
            </Text>
            <Text className={styles.quickStartScreenDescription} block={true}>
              <FormattedMessage
                id="EditOAuthClientScreen.quick-start-screen.question"
                values={{ applicationType: client?.name ?? "" }}
              />
            </Text>
            <QuickStartFrameworkList
              applicationType={client?.x_application_type}
              showOpenTutorialLabelWhenHover={true}
            />
            <div className={styles.quickStartScreenButtons}>
              <PrimaryButton onClick={onNextButtonClick}>
                <FormattedMessage id="next" />
              </PrimaryButton>
            </div>
          </Widget>
        </ScreenContent>
      </ScreenLayoutScrollView>
    );
  };

const EditOAuthClientScreen: React.FC = function EditOAuthClientScreen() {
  const { appID, clientID } = useParams() as {
    appID: string;
    clientID: string;
  };
  const { renderToString } = useContext(Context);
  const form = useAppConfigForm({ appID, constructFormState, constructConfig });
  const { setState, save, isUpdating } = form;
  const navigate = useNavigate();
  const [isRemoveDialogVisible, setIsRemoveDialogVisible] = useState(false);
  const { themes } = useSystemConfig();
  const [searchParams] = useSearchParams();
  const isQuickScreenVisible = useMemo(() => {
    const quickstart = searchParams.get("quickstart");
    return quickstart === "true";
  }, [searchParams]);

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: renderToString("EditOAuthClientScreen.delete-client-dialog.title"),
      subText: renderToString(
        "EditOAuthClientScreen.delete-client-dialog.description"
      ),
    };
  }, [renderToString]);

  const showDialogAndSetRemoveClientByID = useCallback(() => {
    setState((state) => ({ ...state, removeClientByID: clientID }));
    setIsRemoveDialogVisible(true);
  }, [setIsRemoveDialogVisible, setState, clientID]);

  const dismissDialogAndResetRemoveClientByID = useCallback(() => {
    setIsRemoveDialogVisible(false);
    // It is important to reset the removeClientByID
    // Otherwise the next save will remove the oauth client
    setState((state) => ({ ...state, removeClientByID: undefined }));
  }, [setIsRemoveDialogVisible, setState]);

  const onConfirmRemove = useCallback(() => {
    save().then(
      () => {
        navigate("./../..", { replace: true });
      },
      () => {
        dismissDialogAndResetRemoveClientByID();
      }
    );
  }, [save, navigate, dismissDialogAndResetRemoveClientByID]);
  const primaryItems: ICommandBarItemProps[] = useMemo(
    () => [
      {
        key: "remove",
        text: renderToString("EditOAuthClientScreen.delete-client.label"),
        iconProps: { iconName: "Delete" },
        theme: themes.destructive,
        onClick: showDialogAndSetRemoveClientByID,
      },
    ],
    [renderToString, showDialogAndSetRemoveClientByID, themes.destructive]
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (isQuickScreenVisible) {
    return <OAuthQuickStartScreenContent form={form} clientID={clientID} />;
  }

  return (
    <FormContainer form={form} primaryItems={primaryItems}>
      <EditOAuthClientContent form={form} clientID={clientID} />
      <Dialog
        hidden={!isRemoveDialogVisible}
        dialogContentProps={dialogContentProps}
        modalProps={{ isBlocking: isUpdating }}
        onDismiss={dismissDialogAndResetRemoveClientByID}
      >
        <DialogFooter>
          <ButtonWithLoading
            theme={themes.actionButton}
            loading={isUpdating}
            onClick={onConfirmRemove}
            disabled={!isRemoveDialogVisible}
            labelId="confirm"
          />
          <DefaultButton
            onClick={dismissDialogAndResetRemoveClientByID}
            disabled={isUpdating || !isRemoveDialogVisible}
          >
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
    </FormContainer>
  );
};

export default EditOAuthClientScreen;
