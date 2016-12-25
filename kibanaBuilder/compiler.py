import os
import argparse
import json

parser = argparse.ArgumentParser(description='Will read one directory to build the kibana import file in another directory')
parser.add_argument('--source', help='Source folder where to find the json folders and files.')
parser.add_argument('--dest', help='Destination folder where to send resolved json files.')
args = parser.parse_args()

def parseFolder(thePath):
    globalData = []
    try:
        dirList = os.listdir(thePath)
    except FileNotFoundError:
        print('ERROR - We have tried to explore an invalid folder [%s]' % thePath)
        exit(1)
    for elt in dirList:
        newPath = thePath + elt
        if os.path.isdir(newPath) == True:
            newPath += '/'
            #print("dir is : %s" % newPath)
        if elt.split('.')[-1] == 'json':
            newData = handleJsonFile(newPath)
            globalData.append(newData)
            print("This is a file: %s" % elt)
        else:
            print("This is a folder: %s" % elt)
            globalData += parseFolder(newPath)
    return globalData

def handleJsonFile(thePath):
    type = thePath.split('/')[-2]
    print("type of kib is %s" % type)
    with open(thePath) as data_file:
        data = json.load(data_file)
        print(data)
        newData = {"_id":data['title'],"_type":type,"_source":data}
        return newData

if args.source[-1] != '/':
    args.source += '/'


globalData = parseFolder(args.source)
with open(args.dest, 'w') as outfile:
    json.dump(globalData,outfile)
