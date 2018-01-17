import json
import base64
from datetime import datetime


def date_convert(osquery_date_string):
    old_date = osquery_date_string.replace(" UTC", "")
    old_date_format = "%a %b %d %H:%M:%S %Y"
    new_date = datetime.strptime(old_date, old_date_format)
    new_date_string = new_date.strftime("%Y-%m-%dT%H:%M:%S")
    return new_date_string


def lambda_handler(event, context):
    output = []
    succeeded = 0
    failed = 0
    for record in event['records']:
        #print("original record {}".format(record))
        payload=base64.b64decode(record['data'])
        p = json.loads(payload.decode('utf-8'))
        calendar_time = date_convert(p['calendarTime'])
        if calendar_time:
            p['calendarTime'] = calendar_time
            s = json.dumps(p)
            sb = s.encode('utf-8')
            b64 = str(base64.b64encode(sb))
            output_rec = {
                'recordId': record['recordId'],
                'result': 'Ok',
                'data': base64.standard_b64encode(json.dumps(p))
            }
            succeeded += 1
            #print(output_rec)
        else:
            output_rec = {
                'recordId': record['recordId'],
                'result': 'ProcessingFailed',
                'data': record['data']
            }
            failed += 1
        output.append(output_rec)
    print('Processing completed.  Successful records {}, Failed records {}.'.format(succeeded,
                                                                                    failed))
    return {'records': output}