const $bundlePage = $('#bundle-page');

if ($bundlePage.length > 0) {

    $.ajax({
        type: "GET",
        url: '/bundles/' + $bundlePage.attr('data-id') + '/prices.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            if (data === null) {
                data = [];
            }

            Highcharts.chart('prices-chart', $.extend(true, {}, defaultChartOptions, {
                legend: {
                    enabled: false,
                },
                tooltip: {
                    formatter: function () {
                        return this.y.toLocaleString() + '% discount on ' + moment(this.x).format("dddd DD MMM YYYY @ HH:mm");
                    }
                },
                xAxis: {
                    labels: {
                        step: 1,
                        formatter: function () {
                            return moment(this.value).format("Do MMM YY");
                        },
                    },
                },
                yAxis: {
                    title: {
                        text: ''
                    },
                    type: 'linear',
                    max: 100,
                    min: 0,
                    allowDecimals: false,
                    labels: {
                        formatter: function () {
                            return this.value + '%';
                        },
                    },
                },
                series: [
                    {
                        type: 'line',
                        name: 'Price',
                        step: 'left',
                        color: '#28a745',
                        data: data,
                    }
                ],
            }));
        },
    });
}
