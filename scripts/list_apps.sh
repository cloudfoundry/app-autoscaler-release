#!/bin/bash

echo "Fetching all apps with org and space names..."

NEXT_URL="/v3/apps"

while [ "$NEXT_URL" != "null" ]; do
  RESPONSE=$(cf curl "$NEXT_URL")

  echo "$RESPONSE" | jq -c '.resources[]' | while read -r app; do
    APP_NAME=$(echo "$app" | jq -r '.name')
    SPACE_GUID=$(echo "$app" | jq -r '.relationships.space.data.guid')

    # Get space details
    SPACE_INFO=$(cf curl "/v3/spaces/$SPACE_GUID")
    SPACE_NAME=$(echo "$SPACE_INFO" | jq -r '.name')
    ORG_GUID=$(echo "$SPACE_INFO" | jq -r '.relationships.organization.data.guid')

    # Get org details
    ORG_NAME=$(cf curl "/v3/organizations/$ORG_GUID" | jq -r '.name')

    echo "$ORG_NAME / $SPACE_NAME / $APP_NAME"
  done

  NEXT_URL=$(echo "$RESPONSE" | jq -r '.pagination.next.href')
done

