google.charts.load('current', { packages: ['corechart', 'bar', 'table'] });
google.charts.setOnLoadCallback(() => {
    document.getElementById('search').addEventListener('click', async (ev) => {
        const username = document.getElementById('twitter-handle').value;
        //const username = 'elonmusk';
        const resp = await fetch(
            `http://localhost/api/analysis?name=${encodeURI(username)}`
        );
        if (resp.ok) {
            const data = await resp.json();
            // const data = testData;
            createDataSummaryTable(username, data);
            createSentimentScoreChart(username, data.Scores);
            createSentimentDistChart(username, data.ScoreDist);
            createTopicChart(username, data.Topics);
        } else {
            alert('BAD RESPONSE: ' + resp.status + ': ' + (await resp.text()));
        }
    });
});

function createDataSummaryTable(name, data) {
    var tData = new google.visualization.DataTable();
    tData.addColumn('string', 'User');
    tData.addColumn('number', 'Number of Tweets');
    tData.addColumn('number', 'Mean Sentiment');
    tData.addColumn('number', 'Median Sentiment');
    tData.addRow([
        name,
        { v: data.Count, f: data.Count + '' },
        { v: data.MeanScore, f: data.MeanScore + '' },
        { v: data.MedianScore, f: data.MedianScore + '' },
    ]);

    var options = {
        width: '100%',
    };

    // Instantiate and draw the chart.
    var chart = new google.visualization.Table(
        document.getElementById('summary-container')
    );
    chart.draw(tData, options);
}

function createSentimentScoreChart(name, scores) {
    const dataArr = [['Sentiment Score', '# of Entries with Score']];
    scores = scores.map((score) => [score.Score.toFixed(2) + '', score.Count]);
    const data = new google.visualization.arrayToDataTable(
        dataArr.concat(scores)
    );
    const options = {
        title: `Sentiment Scores for ${name}'s Tweets`,
        chartArea: { width: '50%' },
    };
    const chart = new google.visualization.ColumnChart(
        document.getElementById('sentiment-score-container')
    );
    chart.draw(data, options);
}

function createSentimentDistChart(name, scoreDist) {
    const dataArr = [
        ['Sentiment Score Distribution', '# of Entries with Overall Sentiment'],
        ['Negative (S <= -0.5)', scoreDist.Negative],
        ['Neutral  (-0.5 < S < 0.5)', scoreDist.Neutral],
        ['Positive (0.5 <= S)', scoreDist.Positive],
    ];
    const data = new google.visualization.arrayToDataTable(dataArr);
    const options = {
        title: `Sentiment Distribution for ${name}'s posts`,
    };
    const chart = new google.visualization.ColumnChart(
        document.getElementById('sentiment-dist-container')
    );
    chart.draw(data, options);
}

function createTopicChart(name, topics) {
    const dataArr = [['Topic Name', '# of Entries with Topic']];
    topics = topics.map((topic) => [topic.Topic, topic.Count]);
    const data = new google.visualization.arrayToDataTable(
        dataArr.concat(topics)
    );
    const options = {
        title: `Common Topics for ${name}'s Tweets`,
        chartArea: { width: '100%' },
        hAxis: { showTextEvery: 1 },
    };
    const chart = new google.visualization.ColumnChart(
        document.getElementById('topic-container')
    );
    chart.draw(data, options);
}

const testData = {
    Scores: [
        { Score: -0.20000000298023224, Count: 1 },
        { Score: -0.10000000149011612, Count: 5 },
        { Score: 0, Count: 14 },
        { Score: 0.10000000149011612, Count: 4 },
        { Score: 0.20000000298023224, Count: 5 },
        { Score: 0.30000001192092896, Count: 3 },
        { Score: 0.4000000059604645, Count: 7 },
        { Score: 0.5, Count: 3 },
        { Score: 0.6000000238418579, Count: 2 },
        { Score: 0.8999999761581421, Count: 4 },
    ],
    MeanScore: 0.22291666750485697,
    MedianScore: 0.15000000223517418,
    Count: 48,
    ScoreDist: { Negative: 0, Neutral: 39, Positive: 9 },
    Topics: [
        { Topic: '/Computers \u0026 Electronics', Count: 1 },
        { Topic: '/Internet \u0026 Telecom/Web Services', Count: 1 },
        { Topic: '/Science/Engineering \u0026 Technology', Count: 4 },
        {
            Topic:
                '/Business \u0026 Industrial/Aerospace \u0026 Defense/Space Technology',
            Count: 3,
        },
        { Topic: '/Business \u0026 Industrial', Count: 1 },
        { Topic: '/Autos \u0026 Vehicles/Motor Vehicles (By Type)', Count: 1 },
        {
            Topic: '/Computers \u0026 Electronics/Consumer Electronics',
            Count: 1,
        },
    ],
};
