import requests
import response
import json


class Groups:
    def __init__(self, url):
        self.url = url

    def create(self, group, token):
        '''Creates group entity in the database'''
        mf_resp = response.Response()
        http_resp = requests.post(self.url + "/groups", json=group, headers={"Authorization": token})
        if http_resp.status_code != 201:
            mf_resp.error.status = 1
            c = http_resp.status_code
            if c == 400:
                mf_resp.error.message = "Failed due to malformed JSON"
            if c == 401:
                mf_resp.error.message = "Missing or invalid access token provided"
            if c == 409:
                mf_resp.error.message = "Entity already exist"
            if c == 415:
                mf_resp.error.message = "Missing or invalid content type"
            if c == 500:
                mf_resp.error.message = "Unexpected server-side error occurred"
        else:
            location = http_resp.headers.get("location")
            mf_resp.value = location.split('/')[2]
        return mf_resp

    def get(self, groupID, token):
        '''Gets a group entity'''
        mf_resp = response.Response()
        http_resp = requests.get(self.url + "/groups/" + groupID, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            c = http_resp.status_code
            if c == 400:
                mf_resp.error.message = "Failed due to malformed JSON"
            if c == 401:
                mf_resp.error.message = "Missing or invalid access token provided"
            if c == 422:
                mf_resp.error.message = "Database can't process request"
            if c == 500:
                mf_resp.error.message = "Unexpected server-side error occurred"
        else:
            mf_resp.value = http_resp.json()
        return mf_resp

    def get_all(self, query_params, token):
        '''Gets all groups from database'''
        url = self.url + "/groups" + '?' + 'offset=' + query_params['offset'] + '&' + \
            'limit=' + query_params['limit'] + '&' + 'connected=' + query_params['connected']
        mf_resp = response.Response()
        http_resp = requests.get(url, headers={"Authorization": token})
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            c = http_resp.status_code
            if c == 400:
                mf_resp.error.message = "Failed due to malformed channel's ID"
            if c == 401:
                mf_resp.error.message = "Missing or invalid access token provided"
            if c == 404:
                mf_resp.error.message = "Channel does not exist"
            if c == 422:
                mf_resp.error.message = "Database can't process request"
            if c == 500:
                mf_resp.error.message = "Unexpected server-side error occurred"
        else:
            mf_resp.value = json.loads(http_resp.json)
        return mf_resp

    def update(self, group, token):
        '''Updates group entity'''
        http_resp = requests.put(self.url + "/groups/" + group["id"], json=group, headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 200:
            mf_resp.error.status = 1
            c = http_resp.status_code
            if c == 400:
                mf_resp.error.message = "Failed due to malformed JSON"
            if c == 401:
                mf_resp.error.message = "Missing or invalid access token provided"
            if c == 404:
                mf_resp.error.message = "Channel does not exist"
            if c == 415:
                mf_resp.error.message = "Missing or invalid content type"
            if c == 500:
                mf_resp.error.message = "Unexpected server-side error occurred"
        return mf_resp

    def delete(self, groupID, token):
        '''Deletes a group entity from database'''
        http_resp = requests.delete(self.url + "/groups/" + groupID, headers={"Authorization": token})
        mf_resp = response.Response()
        if http_resp.status_code != 204:
            mf_resp.error.status = 1
            c = http_resp.status_code
            if c == 400:
                mf_resp.error.message = "Failed due to malformed channel's ID"
            if c == 401:
                mf_resp.error.message = "Missing or invalid access token provided"
            if c == 500:
                mf_resp.error.message = "Unexpected server-side error occurred"
        return mf_resp
