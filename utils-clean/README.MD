# OBJECTIVE
Process a JSON file that is easier to handle into a JSON file with the structure needed for
Google Datastore.

# FILE DESCRIPTIONS
## utils.js
Provides the method to be used on the other scripts to generate the desired structure for the Google DB

## ./tenant/TenantToDatastore.js
This script should create all data needed for the tenant configuration used on **_common-settings_**, but current it only generates
what is needed as **_contentTutorial_**

## ./journeys/config/JourneysConfigToDatastore.js
This script creates the DB JSON to set up de general configuration of the journeys section, this is the **_config_**
property of **_journeys_** configuration

## ./journeys/icons/iconsToDatastore.js
This script create the DB JSON to set up the icon library that is used on journey's steps, this is the **_icons_**
property of **_journeys_** configuration

## ./journeys/themes/themeToDatastore.js
This script creates de DB JSON to set up journey's themes, this is the **_themes_** 
property of **_journeys_** configuration 

## ./journeys/detail/JourneyDetailToDatastore.js
This script creates one JSON file for each journey with the DB JSON structure to set up each journey 
(**_journeyId_** property value has to match with what is listed on **_journeyList_** property).

It also creates a _journeysList.json_ file that is used to set up the available journey on the system, this is the 
**_journeysList_** property of **_journeys_** configuration
