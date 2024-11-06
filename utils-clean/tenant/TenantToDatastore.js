const fs = require('fs');
const {getI18N, getImages, getStringValue, getColors, buildEntity} = require("../utils");

rawData = fs.readFileSync("./tenant.json")
input = JSON.parse(rawData);

colorList = {
    "overlay" : getStringValue(input.color.overlay),
    "background" : getStringValue(input.color.background)
}
tutorialObj = {
    "i18n": getI18N(input.i18n),
    "images": getImages(input.images),
    "color": buildEntity(colorList),
    "added": getStringValue(new Date().getTime()),
    "cta": getCTAEntity(input.cta)
}


tutorial = buildEntity(tutorialObj)

let jsonData = JSON.stringify(tutorial, null, 2); // Convert JavaScript object to JSON string with pretty print (indentation of 2 spaces)
fs.writeFileSync(`datastore/tutorial.json`, jsonData); // Write JSON data to a file
