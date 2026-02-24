import datetime

def parse_date(date_string: str):
    return datetime.datetime.strptime(date_string, "%Y-%m-%d").date()

def parse_boolean(string: str, default: bool) -> bool:
    if string == None: return default
    if string == "true": return True
    if string == "false": return False
    raise RuntimeError("Could Not Parse Boolean {string}")
