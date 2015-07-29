// app/models/thing.js

var mongoose     = require('mongoose');
var Schema       = mongoose.Schema;

var ThingSchema   = new Schema({
    name: {
        type: String,
        default: '',
        required: 'Please fill Thing name',     //Thing has to have a friendly name
        trim: true
    },
    created: {
        type: Date,
        default: Date.now
    },
    enabled: {
        type: Boolean,
        default: false
    },
    claimable: {
        type: Boolean,
        default: false
    }
});

module.exports = mongoose.model('Thing', ThingSchema);
