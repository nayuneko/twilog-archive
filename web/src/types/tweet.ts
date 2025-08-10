export type TweetResponse = {
    date:   string;
    tweets: TweetResponseTweet[];
}
export type TweetResponseTweet = {
    id: string;
    text: string;
    screen_name: string;
    name?: string;
    created: string;
    retweeted: boolean;
    replied: boolean;
    media?: string[];
    urls?: Urls[];
    hashtags?: string[];
}

type Urls = {
    url: string;
    expanded_url: string;
    display_url: string;
}