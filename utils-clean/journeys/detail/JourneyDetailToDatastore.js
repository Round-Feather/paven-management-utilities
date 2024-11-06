const fs = require('fs');
const path = require('path');
const {getStringValue, getI18N, getImagesValues, getImages, getBooleanValue, getCTAArrayEntity, getArrayEntity
} = require("../../utils");

// let rawData = fs.readFileSync("./output.json")
// let stepsData = fs.readFileSync("./journey-steps.json")
let rawData = fs.readFileSync("./journeys-prod.json")
// let stepsData = fs.readFileSync("./journey-steps-stage.json")
let stepsData = fs.readFileSync("./journey-steps-prod.json")
input = JSON.parse(rawData);


steps = JSON.parse(stepsData)

function deleteAllFilesInDirectory(directory) {
    // read all files in the directory
    fs.readdir(directory, (err, files) => {
        if (err) throw err;

        for (let file of files) {
            // construct full file path
            var filePath = path.join(directory, file);

            // remove the file
            fs.unlink(filePath, err => {
                if (err) throw err;
            });
        }
    });
}

getJourneySteps = (data) => {
    stepsList = data.map(el => {
        return {
            "i18n": getI18N(el.i18n),
            "icons": getStringValue(el.icons),
            "canSkip": getBooleanValue(el.canSkip),
            "url": getStringValue(el.url),
            "type": getStringValue(el.type),
            "id": getStringValue(el.id)
        };
    })

    return getArrayEntity(stepsList)
}
deleteAllFilesInDirectory('./datastore');

getJourney = (data) => {
    journeyStepsData = steps[data.journeyId]
    totalSteps =  journeyStepsData.length
    journey = {
        "journeyId": getStringValue(data.journeyId),
        "i18n": getI18N(data.i18n),
        "images": getImages(data.images),
        "highlighted": getBooleanValue(data.highlighted),
        "done": getBooleanValue(data.done),
        "cta": getCTAArrayEntity(data.cta),
        "theme": getStringValue(data.theme ) || "default",
        "steps": getJourneySteps(journeyStepsData),
        "date": getStringValue(new Date().getTime()),
        "shareAlias": getStringValue(data.shareAlias)
    };
    if (data.video) {
        journey.video = getImagesValues(data.video);
    }

    return journey
}

setTimeout(() => {
    journeyListTemp = [];

    input.map( (el) => {
        config = {
            "properties": getJourney(el)
        }
        let jsonData = JSON.stringify(config, null, 2); // Convert JavaScript object to JSON string with pretty print (indentation of 2 spaces)
        fs.writeFileSync(`./datastore/${el.journeyId}.json`, jsonData); // Write JSON data to a file
        journeyListTemp.push(getStringValue(el.journeyId))
    })
    journeyList = {
        "values": journeyListTemp
    }
    fs.writeFileSync(`./datastore/journeysList.json`, JSON.stringify(journeyList, null, 2)); // Write JSON data to a file
}, 2000)
