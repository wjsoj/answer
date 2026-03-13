#!/bin/bash
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Remove old i18n files to force refresh from embedded resources
if [ -d /data/i18n ]; then
  echo "Removing old i18n files..."
  rm -rf /data/i18n
fi

# Create necessary directories if they don't exist
mkdir -p /data/conf /data/cache /data/uploads

# Check if config file exists
if [ ! -f /data/conf/config.yaml ]; then
  echo "Config file not found, initializing..."
  /usr/bin/answer init -C /data/

  # If AUTO_INSTALL is set, the init command will auto-install and exit
  # Otherwise, it will start the web installer which we don't want in Docker
  if [ -z "$AUTO_INSTALL" ]; then
    echo "ERROR: Config file not created. Please set AUTO_INSTALL environment variables."
    exit 1
  fi
else
  # Config exists but i18n was removed - reinstall i18n bundle
  echo "Reinstalling i18n bundle..."
  /usr/bin/answer i18n -t /data/i18n/
fi

# Run the application
/usr/bin/answer run -C /data/
