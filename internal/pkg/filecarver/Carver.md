#### sauce -> https://github.com/mwielgoszewski/doorman/issues/120

As of osquery version 2.4.5 osquery has the capability to pull, or "carve", files and directories from remote hosts. The design for what a backend looks like is two endpoints, one for beginning the carve session on the backend and a second for receiving the blocks associated with a carves data.

A simple example of how this is done is as follows, taken from our generic test_http_server.py file:

    # Initial endpoint, used to start a carve request
    def start_carve(self, request):
        # The osqueryd agent expects the first endpoint to return a 'session id' through
        # which they'll communicate in future POSTs. We use this internally to connect
        # the request to the person who requested the carve, and to prepare space for the
        # data.
        sid = ''.join(random.choice(string.ascii_uppercase + string.digits) for _ in range(10))

        # The Agent will send up the total number of expected blocks, the size of each block,
        # the size of the carve overall, the carve GUID to identify this specific carve. We
        # check all of these numbers against predefined maximums to ensure that agents aren't
        # able to DOS our endpoints, and that carves are a reasonable size.
        FILE_CARVE_MAP[sid] = {
            'block_count': int(request['block_count']),
            'block_size': int(request['block_size']),
            'blocks_received' : {},
            'carve_size': int(request['carve_size']),
            'carve_guid': request['carve_id'],
        }

        # Lastly we let the agent know that the carve is good to start, and send the session id back
        self._reply({'session_id' : sid})


    # Endpoint where the blocks of the carve are received, and susequently reassembled.
    def continue_carve(self, request):
        # First check if we have already received this block
        if request['block_id'] in FILE_CARVE_MAP[request['session_id']]['blocks_received']:
            return

        # Store block data to be reassembled later
        FILE_CARVE_MAP[request['session_id']]['blocks_received'][int(request['block_id'])] = request['data']

        # Are we expecting to receive more blocks?
        if len(FILE_CARVE_MAP[request['session_id']]['blocks_received']) < FILE_CARVE_MAP[request['session_id']]['block_count']:
            return

        # If not, let's reassemble everything
        out_file_name = FILE_CARVE_DIR+FILE_CARVE_MAP[request['session_id']]['carve_guid']

        # Check the first four bytes for the zstd header. If not no
        # compression was used, it's a generic .tar
        if (base64.standard_b64decode(FILE_CARVE_MAP[request['session_id']]['blocks_received'][0])[0:4] == b'\x28\xB5\x2F\xFD'):
            out_file_name +=  '.zst'
        else:
            out_file_name +=  '.tar'
        f = open(out_file_name, 'wb')
        for x in range(0, FILE_CARVE_MAP[request['session_id']]['block_count']):
            f.write(base64.standard_b64decode(FILE_CARVE_MAP[request['session_id']]['blocks_received'][x]))
        f.close()
        debug("File successfully carved to: %s" % out_file_name)
        FILE_CARVE_MAP[request['session_id']] = {}