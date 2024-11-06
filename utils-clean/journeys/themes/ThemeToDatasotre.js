const fs = require('fs');
const { getStringValue, buildEntity} = require('../../utils');

let rawData = fs.readFileSync("./themeList.json")
input = JSON.parse(rawData);

list = []

getTheme = (key, data) => {
    let colors = {}
    for (let [key, value] of Object.entries(data)) {
        colors[key] = getStringValue(value);
    }
    return buildEntity(colors);
}
output = {}
list = {}
for (let [key, value] of Object.entries(input)) {
    list[key]=(getTheme(key,value))
}

output = {
    "properties": list
}

let jsonData = JSON.stringify(output, null, 2); // Convert JavaScript object to JSON string with pretty print (indentation of 2 spaces)
fs.writeFileSync(`datastore/themes.json`, jsonData); // Write JSON data to a file
