import os
import sys
import requests
import datetime
import warnings
from dateutil import parser
from dotenv import load_dotenv
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
import models

# Script for getting information from the 540 race sheet and parsing it into a JSON for syncing purposes

load_dotenv()

sheets_api_key = os.getenv("GSHEETS_API_KEY")
recordsDataRange = "A6:G"
dateCell = "K3"

spreadsheetId = "1EjxjQhMod_Bgyz2RirtI4h06lbOQKux2N-XhFzCQG_k"
ip = "http://localhost"
port = ":3000"

# Output path
output_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "syncjson/540update.json")

# Importable JSON instance
json_out = models.ImportableJSON()

class Record540:
    def __init__(self, player_name=None, smo_estimate=None, sm3dw_estimate=None, sm64_estimate=None, smg1_estimate=None, sms_estimate=None, smg2_estimate=None):
        self.player = player_name
        self.smo_estimate = smo_estimate
        self.sm3dw_estimate = sm3dw_estimate
        self.sm64_estimate = sm64_estimate
        self.smg1_estimate = smg1_estimate
        self.sms_estimate = sms_estimate
        self.smg2_estimate = smg2_estimate

    def __str__(self):
        return f"SMO: {self.smo_estimate}\nSM3DW: {self.sm3dw_estimate}\nSM64: {self.sm64_estimate}\nSMG1: {self.smg1_estimate}\nSMS: {self.sms_estimate}\nSMG2: {self.smg2_estimate}"

    def SetName(self, name):
        self.player = name

    def SetSMOEstimate(self, estimate):
       self.smo_estimate = estimate

    def SetSM3DWEstimate(self, estimate):
       self.sm3dw_estimate = estimate

    def SetSM64Estimate(self, estimate):
        self.sm64_estimate = estimate

    def SetSMG1Estimate(self, estimate):
        self.smg1_estimate = estimate

    def SetSMSEstimate(self, estimate):
        self.sms_estimate = estimate

    def SetSMG2Estimate(self, estimate):
        self.smg2_estimate = estimate

def main():
    # Get values in the range of records
    r = requests.get(f"https://sheets.googleapis.com/v4/spreadsheets/{spreadsheetId}/values/{recordsDataRange}?key={sheets_api_key}")
    recordsResJson = r.json()

    # Get value of the cell we're using to parse date
    r = requests.get(f"https://sheets.googleapis.com/v4/spreadsheets/{spreadsheetId}/values/{dateCell}?key={sheets_api_key}")
    dateResJson = r.json()

    # Parse date of this race to make sure it's valid
    currentDate = datetime.date.today().strftime('%Y-%m-%d')
    raceDate = dateResJson['values'][0][0]
    
    try:
        raceDate = parser.parse(raceDate, fuzzy=True).strftime('%Y-%m-%d')
    except parser.ParserError:
        print(f"Error: {raceDate} cannot be parsed as a valid date")
        return

    # Make sure this race is happening in the future
    if (raceDate < currentDate):
        print(f"{raceDate} has already passed")
        json_out.Export(output_path)
        return

    # Get existing records and current sheet records
    existingRecords = getExistingRecords(raceDate)
    sheetRecords = getSheetRecords(recordsResJson)
    parseSignups(sheetRecords, existingRecords, raceDate)

    json_out.Export(output_path)

# Takes sheets response json and returns a dict of records in the format {twitchName: record}
def getSheetRecords(responseJSON):
    sheetRecords = {}
    for item in responseJSON["values"]:
        if len(item) == 0:
            break
        newRec = Record540(item[0].lower(), item[1], item[2], item[3], item[4], item[5], item[6])
        sheetRecords[newRec.player] = newRec

    return sheetRecords

# Gets the records for this 540 race
def getExistingRecords(date):
    # First make sure this race exists
    endpoint = f"{ip}{port}/races?on={date}&category=540"
    raceRes = requests.get(endpoint)
    raceRes = raceRes.json()

    # Error with this request
    if raceRes["success"] == False:
        raise Exception(raceRes["error"])

    # If there are no races, add this race
    if len(raceRes["races"]) == 0:
        json_out.NewRace("540", date, start_time="13:00:00")
        return {} # No races means no records

    # If there are races, get this race's ID and get the records from it
    raceID = int(raceRes["races"][0]["id"])
    
    # Get records for this race
    existingRecords = {}
    url = ip + port + "/records/runs?race_id="+str(raceID)
    while True:
        req_records = requests.get(url)
        req_records = req_records.json()
        if req_records["success"] == False:
            raise Exception(req_records["error"])
    
        for run in req_records["runs"]:
            # Get twitch name for this player
            # I should probably fix this in the backend instead of needing to query each of these, but this script runs like every few minutes so for NOW it's fine
            playerName = run["player_name"]
            url = ip + port + "/players?player_name="+playerName
            r = requests.get(url)
            r = r.json()
            if r["success"] == False:
                raise Exception(r["error"])
            playerName = r["players"][0]["twitch_name"]

            category = run["game_category"]
            estimate = run["estimate"]
            if playerName not in existingRecords:
                existingRecords[playerName] = Record540(playerName)
            if category == "sm64_70":
                existingRecords[playerName].SetSM64Estimate(estimate)
            elif category == "smg1_any%":
                existingRecords[playerName].SetSMG1Estimate(estimate)
            elif category == "sms_any%":
                existingRecords[playerName].SetSMSEstimate(estimate)
            elif category == "smg2_any%":
                existingRecords[playerName].SetSMG2Estimate(estimate)
            elif category == "smo_any%":
                existingRecords[playerName].SetSMOEstimate(estimate)
            elif category == "sm3dw_any%":
                existingRecords[playerName].SetSM3DWEstimate(estimate)

        if req_records["meta"]["next_url"] == None:
            break
        url = ip + port + req_records["meta"]["next_url"]
    
    return existingRecords

def parseSignups(sheetRecords, existingRecords, date):
    # For each sheet record, if it doesn't exist in existing records, add it
    for name, record in sheetRecords.items():
        if name not in existingRecords:
            # Check if they exist in the database. Add them if not.
            playerRes = requests.get(ip+port+"/players?twitch_name="+name)
            playerRes = playerRes.json()
            if playerRes["success"] == False:
                warnStr = playerRes["error"] + ": " + name
                warnings.warn(warnStr)
                continue
            if len(playerRes["players"]) == 0:
                json_out.NewPlayer(name)
            
            # Add new record for this player
            newRuns = [
                models.Run("sm64_70", record.sm64_estimate),
                models.Run("smg1_any%", record.smg1_estimate),
                models.Run("sms_any%", record.sms_estimate),
                models.Run("smg2_any%", record.smg2_estimate),
                models.Run("smo_any%", record.smo_estimate),
                models.Run("sm3dw_any%", record.sm3dw_estimate),
            ]
            json_out.NewRecord(name, date, "540", newRuns)

        # If player IS already on the sheet, check for updates
        else:
            updateRuns = []
            if record.sm64_estimate != existingRecords[name].sm64_estimate:
                updateRuns.append(models.Run("sm64_70", record.sm64_estimate))
            if record.smg1_estimate != existingRecords[name].smg1_estimate:
                updateRuns.append(models.Run("smg1_any%", record.smg1_estimate))
            if record.sms_estimate != existingRecords[name].sms_estimate:
                updateRuns.append(models.Run("sms_any%", record.sms_estimate))
            if record.smg2_estimate != existingRecords[name].smg2_estimate:
                updateRuns.append(models.Run("smg2_any%", record.smg2_estimate))
            if record.smo_estimate != existingRecords[name].smo_estimate:
                updateRuns.append(models.Run("smo_any%", record.smo_estimate))
            if record.sm3dw_estimate != existingRecords[name].sm3dw_estimate:
                updateRuns.append(models.Run("sm3dw_any%", record.sm3dw_estimate))
            
            if len(updateRuns) > 0:
                json_out.UpdateRecord(playerName=name, raceDate=date, raceCategory="540", runs=updateRuns)

    # Check if existing record. If they have one that is NOT on the sheet, delete it
    for name, record in existingRecords.items():
        if name not in sheetRecords:
            json_out.DeleteRecord(name, date, "540")

if __name__ == "__main__":
    main()