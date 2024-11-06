const fs = require('fs');
const {getStringValue} = require("../../utils");

rawData = fs.readFileSync("./journeysConfig.json")

input = JSON.parse(rawData);
function getTextEntity(text) {
    return {
        "entityValue": {
            "properties": {
                "text": {
                    "stringValue": text.text.toString()
                },
                "fontColor": {
                    "stringValue": text.fontColor.toString()
                }
            }
        }
    }
}

getArrayStringEntity = (data) => {

    list = data.map( el => {
        return {
            "stringValue": el.toString(),
        }
    })

    return {
        "arrayValue": {
            "values": list
        }
    }
}

getArray = (data) => {
    list = data.map( el => {
        return {
            "stringValue": el.toString(),
        }
    })

    return {
            "values": list
    }
}

getArrayEntity = (data) => {
    entities = data.map( el => {
        return {
            "entityValue": {
                "properties": {
                    "type": {
                        "stringValue": el.type
                    },
                    "action": {
                        "stringValue": el.action
                    },
                    "backgroundColor": {
                        "stringValue": el.backgroundColor
                    },
                    "fontColor": {
                        "stringValue": el.fontColor
                    },
                    "text": {
                        "stringValue": el.text
                    },
                    "textDesktop": {
                        "stringValue": el.textDesktop
                    },
                    "textTablet": {
                        "stringValue": el.textTablet
                    },
                    "disabledBackgroundColor": {
                        "stringValue": el.disabledBackgroundColor
                    },
                    "disabledFontColor": {
                        "stringValue": el.disabledFontColor
                    },
                }
            }
        }
    })

    return {
        "arrayValue": {
            "values": [
                entities
                ]
        }
    }
}

journeysList =
    getArray(input.journeys)

config = {
    "properties": {
        "i18n": {
            "entityValue": {
                "properties": {
                    "title": getTextEntity(input.i18n.title),
                    "description": getTextEntity(input.i18n.description),
                    "inProgress": getTextEntity(input.i18n.inProgress),
                    "completed": getTextEntity(input.i18n.completed)
                }
            }
        },
        "colors": {
            "entityValue": {
                "properties": {
                    "background":getStringValue(input.colors.background)
                }
            }
        },
        "cta": getArrayEntity(input.cta)
    }
}

let jsonData = JSON.stringify(config, null, 2); // Convert JavaScript object to JSON string with pretty pr (indentation of 2 spaces)
fs.writeFileSync('dataStore/journeysConfigDatastore.json', jsonData); // Write JSON data to a file
