-- Create a tracks table containing submitted Bandcamp track data.
CREATE TABLE IF NOT EXISTS tracks (
    track_id    INTEGER,
    track_title TEXT,
    album_title TEXT,
    band_name   TEXT,
    track_url   TEXT,
    created     INTEGER DEFAULT (unixepoch())
);

-- Index columns on the tracks table we'll be looking up frequently.
CREATE INDEX IF NOT EXISTS tracks_idx ON tracks(track_title, album_title, band_name, track_url);