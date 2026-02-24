import requests
import datetime

archive_location = "/Volumes/LaCie/train/archive"
start_date = datetime.date(2026, 1, 1)
end_date = datetime.date(2026, 2, 1)

def download_file(url, path):
    try:
        response = requests.get(url, stream=True)
        response.raise_for_status()

        file = open(path, "wb")

        for chunk in response.iter_content(chunk_size=8192):
            if chunk:
                file.write(chunk)

        file.close()
    except Exception as error:
        print(error)

current_date = start_date

while current_date <= end_date:
    for hour in range (0,24):
        url = f"https://ilovetrains.co.uk/api/download-darwin-dump?fileName={current_date.year:04}%2F{current_date.month:02}%2F{current_date.day:02}%2F{hour:02}.pport.gz"
        path = f"{archive_location}/{current_date.strftime("%Y-%m-%d")}-{hour:02}.pport.gz"

        download_file(url, path)

    current_date += datetime.timedelta(days=1)