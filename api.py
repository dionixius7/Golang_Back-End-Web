import requests
import pickle
import re
import spacy
import pandas as pd
from spacy.matcher import Matcher
from tika import parser
import nltk
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
from nltk.stem import WordNetLemmatizer
from flask import Flask, request, jsonify
import tensorflow as tf
from tensorflow.keras.preprocessing.text import Tokenizer
from tensorflow.keras.preprocessing.sequence import pad_sequences
import numpy as np
import json
from pdfminer.high_level import extract_text


def clean_data(data):
    data = data.lower()
    # URLs
    data = re.sub("http\S+\s*", " ", data)
    # RT and cc
    data = re.sub("RT|cc", " ", data)
    # hashtags
    data = re.sub("#\S+", "", data)
    # mentions
    data = re.sub("@\S+", " ", data)
    # punctuations and non-ASCII char
    data = re.sub("[^a-zA-Z]", " ", data)
    # extra whitespace
    data = re.sub("\s+", " ", data)
    return data

def tokenize1(text):
    words = re.split("\W+", text)
    return words

def remove_stopwords(text):
    stopword = nltk.corpus.stopwords.words("english")
    list_stopword = [
        "skills",
        "education",
        "windows",
        "set",
        "may",
        "january",
        "february",
        "march",
        "april",
        "june",
        "july",
        "august",
        "september",
        "october",
        "november",
        "december",
        "ymcaust",
        "fariabad",
        "b",
        "e",
        "uit",
        "us",
        "other",
        "others",
        "personal",
        "languages",
        "su",
        "o",
        "project",
        "exprience",
        "company",
        "month",
        "team",
        "detail",
        "description",
        "system",
        "application",
        "technology",
        "year",
        "le",
        "ltd",
        "university",
        "j",
        "ge",
        "like",
        "also",
        "timely",
        "per",
        "new",
        "daily",
        "etc",
        "size",
        "level",
        "experience",
        "mumbai",
        "ee",
        "pre",
        "text",
        "co",
        "till",
        "mi",
        "essfully",
        "es",
        "uat",
        "discus",
        "v",
        "ci",
        "no",
        "bachelor",
        "k",
        "v",
        "o",
        "sla",
        "po",
        "hmi",
        "dubai",
        "h",
        "pl",
        "nashik",
        "capgemini",
        "aug",
        "jan",
        "feb",
        "whenever",
        "n",
        "xen",
        "l",
        "u",
    ]

    stopword.extend(list_stopword)
    text = [word for word in text if word not in stopword]
    return text

def tokenize2(text):
    stop_words = set(stopwords.words("english"))
    words = word_tokenize(text)
    words = [
        word.lower()
        for word in words
        if word.lower() not in stop_words and word.isalpha()
    ]
    return words

def lemmatize_word(word):
    lemmatizer = WordNetLemmatizer()
    return [lemmatizer.lemmatize(w) for w in word]

def concat(lst):
    sentence = " ".join(lst)
    return sentence

import os

app = Flask(__name__)

@app.route("/predict", methods=["POST"])  # Ini untuk bisa nampilin nama pekerjaan
def predict():
    if 'document' not in request.files:
        # Handle the case where no file was uploaded
        return "No file part"
    response = request.files['document']
    if response.filename == '':
        # Handle the case where the file has no name
        return "No selected file"

    print("Checkpoint 1: URL SUDAH MASUK")

    file_path = "./file.pdf"
    if response:
        response.save(file_path)
    text = extract_text(file_path)
    text = text.replace("\n", " ")
    text = text.replace("[^a-zA-Z0-9]", " ")
    re.sub("\W+", "", text)
    text = text.lower()

    print("Checkpoint 2: CV UDAH BERSIH")

    # ========================================================================================
    # Praproses Data -> mulai extract cv sampai data preparation (tokenizer, sequence, padded)
    # ==========================================================================================
    # #==========================================================================================
    # # Get the keywords
    Keywords = [
        "education",
        "summary",
        "personal profile",
        "work background",
        "qualifications",
        "experience",
        "achievements",
        "projects",
        "skills",
        "soft skills",
        "hard skills",
    ]

    content = {}
    indices = []
    keys = []

    for key in Keywords:
        try:
            content[key] = text[text.index(key) + len(key) :]
            indices.append(text.index(key))
            keys.append(key)
        except:
            pass
    # Sorting the index
    zipped_lists = zip(indices, keys)
    sorted_pairs = sorted(zipped_lists)

    tuples = zip(*sorted_pairs)
    indices, keys = [list(tuple) for tuple in tuples]
    # ==========================================================================================
    # Remove redundant parts
    parsed_content = {}
    content = []
    for idx in range(len(indices)):
        if idx != len(indices) - 1:
            content.append(text[indices[idx] : indices[idx + 1]])
        else:
            content.append(text[indices[idx] :])

    for i in range(len(indices)):
        parsed_content[keys[i]] = content[i]
    # ==========================================================================================
    df = pd.DataFrame(
        {"keys": keys, "parsed_content": [parsed_content[key] for key in keys]}
    )

    # Extract 'skills' dan 'hard skill' rows

    filtered_df = df[df["keys"].isin(["skills", "hard skill"])]

    data = filtered_df

    print("Checkpoint 3: CV SUDAH KE EKSTRAK SKILL")
    # #==========================================================================================
    # # Clean Data
    filtered_df["parsed_content"] = filtered_df["parsed_content"].apply(clean_data)
    data = filtered_df
    df = pd.DataFrame(data)

    # Convert df to text using to_string()
    text_output = df.to_string(index=False)

    filtered_df["parsed_content_split"] = df["parsed_content"].apply(
        lambda x: tokenize1(x)
    )
    filtered_df["parsed_content_split_stopwords"] = filtered_df["parsed_content"].apply(
        lambda x: tokenize2(x)
    )
    lemmatizer = WordNetLemmatizer()
    filtered_df["parsed_content_split_stopwords_lemma"] = filtered_df[
        "parsed_content_split_stopwords"
    ].apply(lambda x: [lemmatizer.lemmatize(word) for word in x])

    filtered_df["join_words"] = filtered_df[
        "parsed_content_split_stopwords_lemma"
    ].apply(concat)
    new_filtered_df = filtered_df[["keys", "join_words"]]

    data = new_filtered_df
    df = pd.DataFrame(data)

    # convert df to text using to_string()
    text_output = df.to_string(index=False)
    text_output

    data = new_filtered_df
    pd.set_option("display.max_colwidth", 10000)
    text_output = data["join_words"].to_string(index=False)

    print("Checkpoint 4: DATA UDAH SAMPAI SEBELUM DATA PREPARATION")

    # ==========================================================================================
    # Data Preparation

    text_output = tokenize1(text_output)
    text_output = remove_stopwords(text_output)
    text_output = lemmatize_word(text_output)
    text_output = concat(text_output)

    with open(
        r"D:\BANGKITACADEMY2023\lat-ml-skripsi\Klasifikasi-pekerjaan\tokenizer.pickle",
        "rb",
    ) as handle:
        my_tokenizer = pickle.load(handle)

    sequence_output = my_tokenizer.texts_to_sequences([text_output])
    padded_output = pad_sequences(sequence_output)

    # Prediksi
    prediction = model.predict(padded_output)

    print("Checkpoint 5: SUDAH BERHASIL DI PREDIKSI")

    kategori = [
        "Business Analyst",
        "Data Scientist",
        "DevOps Engineer",
        "Java Developer",
        "Operations Manager",
        "Web Designer",
    ]

    prediction_kategori = [kategori[np.argmax(row)] for row in prediction]
    # json_hasil = {"hasil_prediksi": prediction_kategori[0], "text_output": text_output}
    json_hasil = {"hasil_prediksi": prediction_kategori[0]}
    #hapus data
    #submitted_data.pop()
    if os.path.exists(file_path):
        os.remove(file_path)
        print("File berhasil dihapus")

    return jsonify(json_hasil)

if __name__ == "__main__":
    nltk_data_path = "nltk_data"
    nltk.data.path.append(nltk_data_path)
    print("Checkpoint: NLTK berhasil")
    model_use = "D:\BANGKITACADEMY2023\lat-ml-skripsi\Klasifikasi-pekerjaan\classjob.h5"
    model = tf.keras.models.load_model(model_use)
    app.run(host="localhost", port="8080")
