google.charts.load('current', { packages: ['corechart', 'bar', 'table'] });
google.charts.setOnLoadCallback(() => {
    document.getElementById('search').addEventListener('click', async (ev) => {
        startLoading()
        const username = document.getElementById('twitter-handle').value;
        //const username = 'elonmusk';
        console.log(`http://localhost/api/analyse?name=${encodeURI(username)}`)
        const resp = await fetch(
            `http://localhost/api/analyse?name=${encodeURI(username)}`
        );
        if (resp.ok) {
            document.getElementById("loading-indicator-container").classList.remove("loader")
            const data = await resp.json();
            console.log(data)
            if (data.UserID) {
                let positiveCount = data.PositiveTweets
                let negativeCount = data.NegativeTweets
                let averageSentiment = data.AverageScore * 100
                createPositiveNegativeTable(positiveCount, negativeCount)
                createSentimentPieChart(positiveCount, negativeCount)
                createAverageSentimentChart(averageSentiment)
            } else {
                alert(data.Message)
            }
            // const data = testData;
            // createDataSummaryTable(username, data);
            // createSentimentScoreChart(username, data.Scores);
            // createSentimentDistChart(username, data.ScoreDist);
            // createTopicChart(username, data.Topics);
        } else {
            alert('BAD RESPONSE: ' + resp.status + ': ' + (await resp.text()));
        }
    });
});

function createPositiveNegativeTable(positveCount, negativeCount) {
   let tData = google.visualization.arrayToDataTable([
       ['Sentiment', 'Count', { role: 'style' }],
       ['Positive', positveCount, 'green'],
       ['Negative', negativeCount, 'red'],
   ])

    let view = new google.visualization.DataView(tData)
    view.setColumns([0, 1,
        { calc: "stringify",
            sourceColumn: 1,
            type: "string",
            role: "annotation" },
        2]);
    let options = {
        title: "Sentiment of tweets",
        bar: {groupWidth: "95%"},
        legend: { position: "none" },
    };
    let chart = new google.visualization.ColumnChart(document.getElementById("sentiment-counts-container"));
    chart.draw(view, options);
}

function createSentimentPieChart(positiveCount, negativeCount) {
    let data = google.visualization.arrayToDataTable([
        ['Sentiment', 'Count', { role: 'style' }],
        ['Positive', positiveCount, 'green'],
        ['Negative', negativeCount, 'red']
    ])

    let options = {
        title: "Tweet Sentiments",
        slices: {
            0: { color: 'green' },
            1: { color: 'red' }
        }
    }

    let chart = new google.visualization.PieChart(document.getElementById("sentiment-pie-chart-wrapper"))
    chart.draw(data, options)
}

function createAverageSentimentChart(averageSentiment) {
    let data = google.visualization.arrayToDataTable([
        ['Sentiment', 'Average'],
        ['Average Sentiment', averageSentiment]
    ])

    let view = new google.visualization.DataView(data)
    // view.setColumns([0, 1,
    //     { calc: "stringify",
    //         sourceColumn: 1,
    //         type: "string",
    //         role: "annotation" },
    //     2]);

    let options = {
        title: "Average Tweet Sentiment",
        bar: {groupWidth: "95%"},
        legend: { position: "none" },
        hAxis: {
            maxValue: 100,
            minValue: 0,
        }
    };
    let chart = new google.visualization.BarChart(document.getElementById("average-sentiment-wrapper"));
    chart.draw(view, options);
}

const startLoading = () => {
    document.getElementById("sentiment-counts-container").innerHTML = ""
    document.getElementById("sentiment-pie-chart-wrapper").innerHTML = ""
    document.getElementById("average-sentiment-wrapper").innerHTML = ""
    document.getElementById("loading-indicator-container").classList.add("loader")
}