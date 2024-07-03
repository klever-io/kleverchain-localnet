import re

def get_pem_base64(filepath):
      with open(filepath, 'r') as file:
        lines = file.readlines()

        base64_data = ''.join(line.strip() for line in lines if not line.startswith('-----'))

        return base64_data
      
def get_pem_header(filepath):
    with open(filepath, 'r') as file:
        line = file.read()

    match = re.search(r'-----BEGIN PRIVATE KEY for (.*?)-----', line)
    if match:
        return match.group(1)
    
    return ''