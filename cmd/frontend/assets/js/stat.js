const $statPage = $('#stat-page');

if ($statPage.length > 0) {

    $.ajax({
        type: "GET",
        url: '/' + $statPage.attr('data-stat-type') + '/' + $statPage.attr('data-stat-id') + '/time.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            if (data === null) {
                data = [];
            }

            const yAxis = {
                allowDecimals: false,
                title: {
                    text: ''
                },
                labels: {
                    enabled: false
                },
            };

            Highcharts.chart('stat-chart', $.extend(true, {}, defaultChartOptions, {
                yAxis: [
                    yAxis,
                    yAxis,
                    yAxis,
                    yAxis,
                    yAxis,
                ],
                tooltip: {
                    formatter: function () {
                        switch (this.series.name) {
                            case 'Apps':
                                return Math.round(this.y).toLocaleString() + ' games with tag on ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'Apps (%)':
                                return this.y.toLocaleString() + '% of games have tag ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'Mean Players':
                                return this.y.toLocaleString() + ' mean max weakly players on ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'Mean Price (' + user.userCurrencySymbol + ')':
                                return user.userCurrencySymbol + ' ' + (this.y / 100).toFixed(2).toLocaleString() + ' mean price on ' + moment(this.key).format("dddd DD MMM YYYY");
                            case 'Mean Review Score':
                                return this.y.toLocaleString() + '% mean review score on ' + moment(this.key).format("dddd DD MMM YYYY");
                        }
                    },
                },
                series: [
                    {
                        name: 'Apps',
                        data: data['max_apps_count'],
                        marker: {symbol: 'circle'},
                        yAxis: 0,
                        visible: false,
                    },
                    {
                        name: 'Apps (%)',
                        data: data['max_apps_percent'],
                        marker: {symbol: 'circle'},
                        yAxis: 1,
                    },
                    {
                        name: 'Mean Players',
                        data: data['max_mean_players'],
                        marker: {symbol: 'circle'},
                        yAxis: 2,
                    },
                    {
                        name: 'Mean Price (' + user.userCurrencySymbol + ')',
                        data: data['max_mean_price_uk'],
                        marker: {symbol: 'circle'},
                        yAxis: 3,
                    },
                    {
                        name: 'Mean Review Score',
                        data: data['max_mean_score'],
                        marker: {symbol: 'circle'},
                        yAxis: 4,
                    },
                ],
            }));
        },
    });
}