import React from "react";
import SubscriptionPlanCard, {
  BasePriceTag,
  CTA,
  CardTagline,
  CardTitle,
  MAURestriction,
  PlanDetailsLine,
  PlanDetailsTitle,
  UsagePriceTag,
} from "./SubscriptionPlanCard";
import { FormattedMessage } from "@oursky/react-messageformat";

export interface SubscriptionEnterprisePlanProps {
  previousPlanName: string | null;
  onClickContactUs?: () => void;
}

export const SubscriptionEnterprisePlan: React.VFC<SubscriptionEnterprisePlanProps> =
  function SubscriptionEnterprisePlan({ previousPlanName, onClickContactUs }) {
    return (
      <SubscriptionPlanCard
        isCurrentPlan={false}
        cardTitle={
          <CardTitle>
            <FormattedMessage id={"SubscriptionScreen.plan-name.enterprise"} />
          </CardTitle>
        }
        cardTagline={
          <CardTagline>
            <FormattedMessage
              id={"SubscriptionPlanCard.plan.tagline.enterprise"}
            />
          </CardTagline>
        }
        basePriceTag={
          <BasePriceTag>
            <FormattedMessage
              id={"SubscriptionPlanCard.plan.price.enterprise"}
            />
          </BasePriceTag>
        }
        mauRestriction={
          <MAURestriction>
            <FormattedMessage
              id={"SubscriptionPlanCard.plan.mau-restriction.enterprise"}
            />
          </MAURestriction>
        }
        usagePriceTags={
          <>
            {
              <UsagePriceTag>
                <FormattedMessage
                  id="SubscriptionPlanCard.sms.north-america"
                  values={{
                    unitAmount: "0.01",
                  }}
                />
              </UsagePriceTag>
            }
            {
              <UsagePriceTag>
                <FormattedMessage
                  id="SubscriptionPlanCard.sms.other-regions"
                  values={{
                    unitAmount: "0.06",
                  }}
                />
              </UsagePriceTag>
            }
          </>
        }
        cta={<CTA.ContactUs onClick={onClickContactUs} />}
        planDetailsTitle={
          previousPlanName ? (
            <PlanDetailsTitle>
              <FormattedMessage
                id="SubscriptionPlanCard.plan.features.title"
                values={{
                  previousPlan: (
                    <FormattedMessage
                      id={`SubscriptionScreen.plan-name.${previousPlanName}`}
                    />
                  ),
                }}
              />
            </PlanDetailsTitle>
          ) : null
        }
        planDetailsLines={
          <>
            <PlanDetailsLine>
              <FormattedMessage
                id={`SubscriptionPlanCard.plan.features.line.0.enterprise`}
              />
            </PlanDetailsLine>
            <PlanDetailsLine>
              <FormattedMessage
                id={`SubscriptionPlanCard.plan.features.line.1.enterprise`}
              />
            </PlanDetailsLine>
            <PlanDetailsLine>
              <FormattedMessage
                id={`SubscriptionPlanCard.plan.features.line.2.enterprise`}
              />
            </PlanDetailsLine>
            <PlanDetailsLine>
              <FormattedMessage
                id={`SubscriptionPlanCard.plan.features.line.3.enterprise`}
              />
            </PlanDetailsLine>
            <PlanDetailsLine>
              <FormattedMessage
                id={`SubscriptionPlanCard.plan.features.line.4.enterprise`}
              />
            </PlanDetailsLine>
          </>
        }
      />
    );
  };
