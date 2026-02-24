import xml.etree.ElementTree as ET
import io

def fetch_namespace_dictionary(chunk):
    output = {}

    for _, element in ET.iterparse(io.StringIO(chunk), events=("start-ns",)):
        prefix, uri = element
        output[prefix] = uri

    return output