
buildEntity = (data) => {
    return {
        "entityValue": {
            "properties": data
        }
    }
}
getIntValue = (int) => {
    return {
        "integerValue": Number(int)
    }
}
getStringValue = (text) => {
    if (!text) return
    return {
        "stringValue": text.toString()
    }
}
getImagesValues = (imagesObject) => {
    let platformImages = {}
    for (let [key, value] of Object.entries(imagesObject)) {
        if (value) platformImages[key] = getStringValue(value);
    }
    if (Object.keys(platformImages).length === 0) return null

    return {
        "entityValue": {
            "properties": platformImages
        }
    }
}
getCTAEntity = (el) => {
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
                "url": getStringValue(el.url)
            }
        }
    }
}
getCTAArrayEntity = (data) => {
    entities = data.map( el => getCTAEntity(el))

    return {
        "arrayValue": {
            "values": [
                entities
            ]
        }
    }
}
getBooleanValue = (flag) => {
    return {
        "booleanValue": Boolean(flag)
    }
}


getImages = (imagesObject) => {
    let images = {}
    for (let [key, value] of Object.entries(imagesObject)) {
        let platformImages = getImagesValues(value);
        if (platformImages) images[key] = platformImages;
    }
    if (Object.keys(images).length === 0) return
    return {
        "entityValue": {
            "properties": images
        }
    }
}
getColors = (colorsObject) => {
    let colors = {}
    for (let [key, value] of Object.entries(colorsObject)) {
        let colorsList = getArrayStringEntity(value);
        if (colorsList) colors[key] = colorsList;
    }

    return buildEntity(colors)
}
getTextEntity = (text) => {
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
getI18N = (i18nObject) => {
    let properties = {}
    for (let [key, value] of Object.entries(i18nObject)) {
        properties[key] = getTextEntity(value)
    }
    return {
        "entityValue": {
            "properties": properties
        }
    }
}
getArrayStringEntity = (data) => {
    let list = data.map(el => {
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
getArrayEntity = (data) => {

    let entities = data.map(el => buildEntity(el));

    return {
        "arrayValue": {
            "values": [
                entities
            ]
        }
    }
}
getArrayEntityOK = (data) => {

    let entities = data.map(el => buildEntity(el));

    return {
        "arrayValue": {
            "values": entities
        }
    }
}


module.exports = {
    buildEntity,
    getIntValue,
    getStringValue,
    getImagesValues,
    getCTAArrayEntity,
    getBooleanValue,
    getImages,
    getColors,
    getTextEntity,
    getI18N,
    getArrayStringEntity,
    getArrayEntity,
    getArrayEntityOK,
    getCTAEntity
}
