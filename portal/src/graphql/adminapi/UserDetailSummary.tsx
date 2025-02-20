import React from "react";
import cn from "classnames";
import { Link } from "react-router-dom";
import { Persona, PersonaSize, Text, FontIcon } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailSummary.module.css";

interface UserDetailSummaryProps {
  className?: string;
  isAnonymous: boolean;
  formattedName?: string;
  endUserAccountIdentifier: string | undefined;
  profileImageURL: string | undefined;
  profileImageEditable: boolean;
  createdAtISO: string | null;
  lastLoginAtISO: string | null;
}

const UserDetailSummary: React.FC<UserDetailSummaryProps> =
  function UserDetailSummary(props: UserDetailSummaryProps) {
    const {
      isAnonymous,
      formattedName,
      endUserAccountIdentifier,
      profileImageURL,
      profileImageEditable,
      createdAtISO,
      lastLoginAtISO,
      className,
    } = props;
    const { locale } = React.useContext(Context);
    const formatedSignedUp = React.useMemo(() => {
      return formatDatetime(locale, createdAtISO);
    }, [locale, createdAtISO]);
    const formatedLastLogin = React.useMemo(() => {
      return formatDatetime(locale, lastLoginAtISO);
    }, [locale, lastLoginAtISO]);

    return (
      <section className={cn(styles.root, className)}>
        <div className={styles.profilePic}>
          <Persona
            imageUrl={profileImageURL}
            size={PersonaSize.size72}
            hidePersonaDetails={true}
          />
          {profileImageEditable && (
            <Link className={styles.cameraLink} to="./edit-picture">
              <FontIcon className={styles.cameraIcon} iconName="Camera" />
            </Link>
          )}
        </div>
        <div className={styles.lines}>
          {isAnonymous && (
            <Text className={styles.anonymousUserLabel} variant="medium">
              <FormattedMessage id="UsersList.anonymous-user" />
            </Text>
          )}
          <Text className={styles.accountID} variant="medium">
            {endUserAccountIdentifier ?? ""}
          </Text>
          <Text className={styles.formattedName} variant="medium">
            {formattedName ? formattedName : ""}
          </Text>
          <Text variant="small">
            <FormattedMessage
              id="UserDetails.signed-up"
              values={{ datetime: formatedSignedUp ?? "" }}
            />
          </Text>
          <Text variant="small">
            <FormattedMessage
              id="UserDetails.last-login-at"
              values={{ datetime: formatedLastLogin ?? "" }}
            />
          </Text>
        </div>
      </section>
    );
  };

export default UserDetailSummary;
